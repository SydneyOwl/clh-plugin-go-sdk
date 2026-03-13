package clhplugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/SydneyOwl/clh-proto/gen/go/v20260312"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Client struct {
	manifest PluginManifest
	cfg      Config

	connMu sync.RWMutex
	conn   net.Conn

	writeMu sync.Mutex

	pendingMu sync.Mutex
	pending   map[string]chan *pb.PipeEnvelope

	waitCh chan Message

	doneCh chan struct{}
	stopCh chan struct{}
	endMu  sync.Once

	reqSeq atomic.Uint64

	connected atomic.Bool
	closed    atomic.Bool

	registerResp RegisterResponse
}

func NewClient(manifest PluginManifest, opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	if err := normalizeManifest(&manifest); err != nil {
		return nil, err
	}

	c := &Client{
		manifest: manifest,
		cfg:      cfg,
		pending:  make(map[string]chan *pb.PipeEnvelope),
		waitCh:   make(chan Message, cfg.WaitBufferSize),
		doneCh:   make(chan struct{}),
		stopCh:   make(chan struct{}),
	}
	return c, nil
}

func normalizeManifest(manifest *PluginManifest) error {
	if manifest == nil {
		return ErrInvalidManifest
	}
	if manifest.UUID == "" || manifest.Name == "" || manifest.Version == "" {
		return fmt.Errorf("%w: uuid/name/version are required", ErrInvalidManifest)
	}
	if manifest.Metadata == nil {
		manifest.Metadata = map[string]string{}
	}
	if manifest.SDKName == "" {
		manifest.SDKName = defaultSDKName
	}
	if manifest.SDKVersion == "" {
		manifest.SDKVersion = defaultSDKVersion
	}
	return nil
}

func (c *Client) Connect(ctx context.Context) (RegisterResponse, error) {
	if c.closed.Load() {
		return RegisterResponse{}, ErrClientClosed
	}
	if c.connected.Load() {
		return c.registerResp, nil
	}

	conn, err := dialPipe(ctx, c.cfg.PipePath)
	if err != nil {
		return RegisterResponse{}, err
	}

	req := toPBManifest(c.manifest)
	if err = writeDelimitedMessage(conn, req); err != nil {
		_ = conn.Close()
		return RegisterResponse{}, err
	}

	resp := &pb.PipeRegisterPluginResp{}
	if err = readDelimitedMessage(conn, resp); err != nil {
		_ = conn.Close()
		return RegisterResponse{}, err
	}

	modelResp := fromPBRegisterResponse(resp)
	if !resp.Success {
		_ = conn.Close()
		if modelResp.Message == "" {
			modelResp.Message = "register failed"
		}
		return modelResp, errors.New(modelResp.Message)
	}

	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()

	c.registerResp = modelResp
	c.connected.Store(true)

	go c.readLoop()
	if c.cfg.HeartbeatInterval > 0 {
		go c.heartbeatLoop()
	}

	return modelResp, nil
}

func (c *Client) IsConnected() bool {
	return c.connected.Load() && !c.closed.Load()
}

func (c *Client) RegisterResponse() RegisterResponse {
	return c.registerResp
}

func (c *Client) WaitMessage(ctx context.Context) (Message, error) {
	if c.closed.Load() && len(c.waitCh) == 0 {
		return Message{}, ErrClientClosed
	}
	select {
	case msg, ok := <-c.waitCh:
		if !ok {
			return Message{}, ErrClientClosed
		}
		return msg, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

func (c *Client) Close(ctx context.Context) error {
	if c.closed.Swap(true) {
		select {
		case <-c.doneCh:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	close(c.stopCh)

	if !c.connected.Load() {
		c.finish()
		return nil
	}

	if c.connected.Load() {
		_ = c.sendAnyMessage(&pb.PipeDeregisterPluginReq{
			Uuid:      c.manifest.UUID,
			Reason:    "client-close",
			Timestamp: nowTimestamp(),
		})
	}

	c.connMu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.connMu.Unlock()

	select {
	case <-c.doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(c.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.sendAnyMessage(&pb.PipeHeartbeat{
				Uuid:      c.manifest.UUID,
				Timestamp: nowTimestamp(),
			})
			if err != nil {
				return
			}
		case <-c.stopCh:
			return
		}
	}
}

func (c *Client) readLoop() {
	defer c.finish()

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		conn, err := c.getConn()
		if err != nil {
			return
		}

		anyMsg := &anypb.Any{}
		if err = readDelimitedMessage(conn, anyMsg); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return
			}
			return
		}

		protoMsg, modelMsg, convErr := fromAnyMessage(anyMsg)
		if env, ok := protoMsg.(*pb.PipeEnvelope); ok && env.Kind == pb.PipeEnvelopeKind_PIPE_ENVELOPE_KIND_RESPONSE {
			c.resolvePending(env)
		}
		if convErr != nil && modelMsg.Kind == InboundKindUnknown && modelMsg.Unknown == nil {
			modelMsg = Message{
				Kind: InboundKindUnknown,
				Unknown: &UnknownMessage{
					TypeURL: anyMsg.GetTypeUrl(),
					Raw:     append([]byte(nil), anyMsg.GetValue()...),
				},
			}
		}
		c.dispatchMessage(modelMsg)

		if _, ok := protoMsg.(*pb.PipeConnectionClosed); ok {
			return
		}
	}
}

func (c *Client) finish() {
	c.endMu.Do(func() {
		c.connected.Store(false)
		c.rejectAllPending()
		close(c.waitCh)
		close(c.doneCh)
	})
}

func (c *Client) dispatchMessage(msg Message) {
	if c.cfg.OnMessage != nil {
		handler := c.cfg.OnMessage
		go handler(msg)
	}

	select {
	case c.waitCh <- msg:
	default:
		select {
		case <-c.waitCh:
		default:
		}
		select {
		case c.waitCh <- msg:
		default:
		}
	}
}

func (c *Client) getConn() (net.Conn, error) {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	if c.conn == nil {
		return nil, ErrNotConnected
	}
	return c.conn, nil
}

func (c *Client) sendAnyMessage(msg proto.Message) error {
	if c.closed.Load() {
		return ErrClientClosed
	}
	if !c.connected.Load() {
		return ErrNotConnected
	}

	conn, err := c.getConn()
	if err != nil {
		return err
	}
	packed, err := anypb.New(msg)
	if err != nil {
		return err
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return writeDelimitedMessage(conn, packed)
}

func (c *Client) nextRequestID() string {
	seq := c.reqSeq.Add(1)
	return fmt.Sprintf("%s-%d-%d", c.manifest.UUID, time.Now().UnixNano(), seq)
}

func (c *Client) resolvePending(resp *pb.PipeEnvelope) {
	key := resp.CorrelationId
	if key == "" {
		key = resp.Id
	}
	if key == "" {
		return
	}

	c.pendingMu.Lock()
	ch := c.pending[key]
	c.pendingMu.Unlock()
	if ch == nil {
		return
	}

	select {
	case ch <- resp:
	default:
	}
}

func (c *Client) rejectAllPending() {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	for k, ch := range c.pending {
		select {
		case ch <- nil:
		default:
		}
		close(ch)
		delete(c.pending, k)
	}
}

func (c *Client) requestRaw(
	ctx context.Context,
	kind EnvelopeKind,
	topic EnvelopeTopic,
	attributes map[string]string,
	payload proto.Message,
	subscription *EventSubscription,
) (*pb.PipeEnvelope, error) {
	if c.closed.Load() {
		return nil, ErrClientClosed
	}
	if !c.connected.Load() {
		return nil, ErrNotConnected
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline && c.cfg.RequestTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.cfg.RequestTimeout)
		defer cancel()
	}

	reqID := c.nextRequestID()
	req := &pb.PipeEnvelope{
		Id:         reqID,
		Kind:       pb.PipeEnvelopeKind(kind),
		Topic:      pb.PipeEnvelopeTopic(topic),
		Success:    true,
		Message:    "request",
		Attributes: map[string]string{},
		Timestamp:  nowTimestamp(),
	}
	for k, v := range attributes {
		req.Attributes[k] = v
	}
	if subscription != nil {
		req.Subscription = toPBEventSubscription(subscription)
	}
	if payload != nil {
		reqPayload, err := anypb.New(payload)
		if err != nil {
			return nil, err
		}
		req.Payload = reqPayload
	}

	respCh := make(chan *pb.PipeEnvelope, 1)
	c.pendingMu.Lock()
	c.pending[reqID] = respCh
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, reqID)
		c.pendingMu.Unlock()
	}()

	if err := c.sendAnyMessage(req); err != nil {
		return nil, err
	}

	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, ErrClientClosed
		}
		return resp, nil
	case <-c.doneCh:
		return nil, ErrClientClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) requestExpectSuccess(
	ctx context.Context,
	kind EnvelopeKind,
	topic EnvelopeTopic,
	attributes map[string]string,
	payload proto.Message,
	subscription *EventSubscription,
) (*pb.PipeEnvelope, error) {
	resp, err := c.requestRaw(ctx, kind, topic, attributes, payload, subscription)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, &RemoteError{
			Topic:         topic,
			Code:          resp.ErrorCode,
			Message:       resp.Message,
			CorrelationID: resp.CorrelationId,
		}
	}
	return resp, nil
}

func decodeResponsePayload(resp *pb.PipeEnvelope, out proto.Message) error {
	if resp.Payload == nil {
		return errors.New("response payload is empty")
	}
	return resp.Payload.UnmarshalTo(out)
}

func (c *Client) QueryServerInfo(ctx context.Context) (ServerInfo, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQueryServerInfo, nil, nil, nil)
	if err != nil {
		return ServerInfo{}, err
	}
	p := &pb.PipeServerInfo{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return ServerInfo{}, err
	}
	return fromPBServerInfo(p), nil
}

func (c *Client) QueryConnectedPlugins(ctx context.Context) (PluginList, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQueryConnectedPlugins, nil, nil, nil)
	if err != nil {
		return PluginList{}, err
	}
	p := &pb.PipePluginList{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return PluginList{}, err
	}
	return fromPBPluginList(p), nil
}

func (c *Client) QueryRuntimeSnapshot(ctx context.Context) (RuntimeSnapshot, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQueryRuntimeSnapshot, nil, nil, nil)
	if err != nil {
		return RuntimeSnapshot{}, err
	}
	p := &pb.PipeRuntimeSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return RuntimeSnapshot{}, err
	}
	return fromPBRuntimeSnapshot(p), nil
}

func (c *Client) QueryRigSnapshot(ctx context.Context) (RigSnapshot, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQueryRigSnapshot, nil, nil, nil)
	if err != nil {
		return RigSnapshot{}, err
	}
	p := &pb.PipeRigStatusSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return RigSnapshot{}, err
	}
	return fromPBRigSnapshot(p), nil
}

func (c *Client) QueryUDPSnapshot(ctx context.Context) (UDPSnapshot, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQueryUDPSnapshot, nil, nil, nil)
	if err != nil {
		return UDPSnapshot{}, err
	}
	p := &pb.PipeUdpStatusSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return UDPSnapshot{}, err
	}
	return fromPBUDPSnapshot(p), nil
}

func (c *Client) QueryQSOQueueSnapshot(ctx context.Context) (QSOQueueSnapshot, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQueryQSOQueueSnapshot, nil, nil, nil)
	if err != nil {
		return QSOQueueSnapshot{}, err
	}
	p := &pb.PipeQsoQueueSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return QSOQueueSnapshot{}, err
	}
	return fromPBQSOQueueSnapshot(p), nil
}

func (c *Client) QuerySettingsSnapshot(ctx context.Context) (SettingsSnapshot, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQuerySettingsSnapshot, nil, nil, nil)
	if err != nil {
		return SettingsSnapshot{}, err
	}
	p := &pb.PipeMainSettingsSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return SettingsSnapshot{}, err
	}
	return fromPBSettingsSnapshot(p), nil
}

func (c *Client) QueryPluginTelemetry(ctx context.Context, pluginUUID string) (PluginTelemetry, error) {
	attrs := map[string]string{}
	if pluginUUID != "" {
		attrs["plugin_uuid"] = pluginUUID
	}
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, EnvelopeTopicQueryPluginTelemetry, attrs, nil, nil)
	if err != nil {
		return PluginTelemetry{}, err
	}
	p := &pb.PipePluginTelemetry{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return PluginTelemetry{}, err
	}
	return fromPBPluginTelemetry(p), nil
}

func (c *Client) SubscribeEvents(ctx context.Context, sub EventSubscription) (EventSubscription, error) {
	resp, err := c.requestExpectSuccess(
		ctx,
		EnvelopeKindCommand,
		EnvelopeTopicCommandSubscribeEvents,
		nil,
		toPBEventSubscription(&sub),
		&sub,
	)
	if err != nil {
		return EventSubscription{}, err
	}
	p := &pb.PipeEventSubscription{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return EventSubscription{}, err
	}
	return fromPBEventSubscription(p), nil
}

func (c *Client) ShowMainWindow(ctx context.Context) error {
	_, err := c.requestExpectSuccess(ctx, EnvelopeKindCommand, EnvelopeTopicCommandShowMainWindow, nil, nil, nil)
	return err
}

func (c *Client) HideMainWindow(ctx context.Context) error {
	_, err := c.requestExpectSuccess(ctx, EnvelopeKindCommand, EnvelopeTopicCommandHideMainWindow, nil, nil, nil)
	return err
}

func (c *Client) OpenWindow(ctx context.Context, window ControllableWindow, asDialog bool) error {
	attrs := map[string]string{
		"window":   string(window),
		"asDialog": fmt.Sprintf("%t", asDialog),
	}
	_, err := c.requestExpectSuccess(ctx, EnvelopeKindCommand, EnvelopeTopicCommandOpenWindow, attrs, nil, nil)
	return err
}

func (c *Client) SendNotification(ctx context.Context, command NotificationCommand) error {
	_, err := c.requestExpectSuccess(
		ctx,
		EnvelopeKindCommand,
		EnvelopeTopicCommandSendNotification,
		nil,
		toPBNotification(command),
		nil,
	)
	return err
}

func (c *Client) ToggleUDPServer(ctx context.Context, enabled *bool) (UDPSnapshot, error) {
	attrs := map[string]string{}
	if enabled != nil {
		attrs["enabled"] = fmt.Sprintf("%t", *enabled)
	}
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindCommand, EnvelopeTopicCommandToggleUDPServer, attrs, nil, nil)
	if err != nil {
		return UDPSnapshot{}, err
	}
	p := &pb.PipeUdpStatusSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return UDPSnapshot{}, err
	}
	return fromPBUDPSnapshot(p), nil
}

func (c *Client) ToggleRigBackend(ctx context.Context, enabled *bool) (RigSnapshot, error) {
	attrs := map[string]string{}
	if enabled != nil {
		attrs["enabled"] = fmt.Sprintf("%t", *enabled)
	}

	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindCommand, EnvelopeTopicCommandToggleRigBackend, attrs, nil, nil)
	if err != nil {
		return RigSnapshot{}, err
	}
	p := &pb.PipeRigStatusSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return RigSnapshot{}, err
	}
	return fromPBRigSnapshot(p), nil
}

func (c *Client) SwitchRigBackend(ctx context.Context, backend RigBackend) (RigSnapshot, error) {
	if backend == "" {
		return RigSnapshot{}, errors.New("backend is required")
	}

	attrs := map[string]string{
		"backend": string(backend),
	}
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindCommand, EnvelopeTopicCommandSwitchRigBackend, attrs, nil, nil)
	if err != nil {
		return RigSnapshot{}, err
	}
	p := &pb.PipeRigStatusSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return RigSnapshot{}, err
	}
	return fromPBRigSnapshot(p), nil
}

func (c *Client) UploadExternalQSO(ctx context.Context, adifLogs string) (Envelope, error) {
	if adifLogs == "" {
		return Envelope{}, errors.New("adifLogs is required")
	}

	resp, err := c.requestExpectSuccess(
		ctx,
		EnvelopeKindCommand,
		EnvelopeTopicCommandUploadExternalQSO,
		map[string]string{"adifLogs": adifLogs},
		nil,
		nil,
	)
	if err != nil {
		return Envelope{}, err
	}
	return fromPBEnvelope(resp), nil
}

// StartRigBackend is kept for compatibility and maps to ToggleRigBackend(enabled=true).
func (c *Client) StartRigBackend(ctx context.Context) (RigSnapshot, error) {
	enabled := true
	return c.ToggleRigBackend(ctx, &enabled)
}

// StopRigBackend is kept for compatibility and maps to ToggleRigBackend(enabled=false).
func (c *Client) StopRigBackend(ctx context.Context) (RigSnapshot, error) {
	enabled := false
	return c.ToggleRigBackend(ctx, &enabled)
}

// RestartRigBackend is kept for compatibility and maps to stop then start.
func (c *Client) RestartRigBackend(ctx context.Context) (RigSnapshot, error) {
	if _, err := c.StopRigBackend(ctx); err != nil {
		return RigSnapshot{}, err
	}
	return c.StartRigBackend(ctx)
}

func (c *Client) TriggerQSOReupload(ctx context.Context, attributes map[string]string) (Envelope, error) {
	if attributes == nil || attributes["qsoIds"] == "" {
		return Envelope{}, errors.New("qsoIds is required")
	}

	resp, err := c.requestExpectSuccess(
		ctx,
		EnvelopeKindCommand,
		EnvelopeTopicCommandTriggerQSOReupload,
		attributes,
		nil,
		nil,
	)
	if err != nil {
		return Envelope{}, err
	}
	return fromPBEnvelope(resp), nil
}

func (c *Client) UpdateSettings(ctx context.Context, patch SettingsPatch) (SettingsSnapshot, error) {
	resp, err := c.requestExpectSuccess(
		ctx,
		EnvelopeKindCommand,
		EnvelopeTopicCommandUpdateSettings,
		nil,
		toPBSettingsPatch(patch),
		nil,
	)
	if err != nil {
		return SettingsSnapshot{}, err
	}
	p := &pb.PipeMainSettingsSnapshot{}
	if err = decodeResponsePayload(resp, p); err != nil {
		return SettingsSnapshot{}, err
	}
	return fromPBSettingsSnapshot(p), nil
}

func (c *Client) RawQuery(
	ctx context.Context,
	topic EnvelopeTopic,
	attributes map[string]string,
	payload proto.Message,
) (Envelope, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindQuery, topic, attributes, payload, nil)
	if err != nil {
		return Envelope{}, err
	}
	return fromPBEnvelope(resp), nil
}

func (c *Client) RawCommand(
	ctx context.Context,
	topic EnvelopeTopic,
	attributes map[string]string,
	payload proto.Message,
	subscription *EventSubscription,
) (Envelope, error) {
	resp, err := c.requestExpectSuccess(ctx, EnvelopeKindCommand, topic, attributes, payload, subscription)
	if err != nil {
		return Envelope{}, err
	}
	return fromPBEnvelope(resp), nil
}
