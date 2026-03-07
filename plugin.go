package pluginsdk

import (
	"bufio"
	"context"
	"fmt"
	plugin "github.com/SydneyOwl/clh-proto/gen/go/v20260307"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net"
	"sync"
	"time"
)

const (
	DefaultHeartbeatInterval = time.Second * 5
	DefaultSdkName           = "clh-plugin-go-sdk"
	DefaultSdkVersion        = "next"
)

type Option func(client *ClhClient) error

func WithHeartbeatInterval(interval time.Duration) Option {
	return func(client *ClhClient) error {
		if interval < time.Second {
			return fmt.Errorf("heartbeat interval should > 1s")
		}

		client.heartbeatInterval = interval
		return nil
	}
}

func WithMetadata(metadata map[string]string) Option {
	return func(client *ClhClient) error {
		if client.pluginCfg.Metadata == nil {
			client.pluginCfg.Metadata = make(map[string]string)
		}
		for k, v := range metadata {
			client.pluginCfg.Metadata[k] = v
		}
		return nil
	}
}

func WithMetadataEntry(key, value string) Option {
	return func(client *ClhClient) error {
		if key == "" {
			return fmt.Errorf("metadata key should not be empty")
		}
		if client.pluginCfg.Metadata == nil {
			client.pluginCfg.Metadata = make(map[string]string)
		}
		client.pluginCfg.Metadata[key] = value
		return nil
	}
}

func WithSdkInfo(name, version string) Option {
	return func(client *ClhClient) error {
		client.pluginCfg.SdkName = name
		client.pluginCfg.SdkVersion = version
		return nil
	}
}

func WithWsjtxSubscription(subscription WsjtxSubscription) Option {
	return func(client *ClhClient) error {
		subCopy := subscription
		if subscription.MessageTypes != nil {
			subCopy.MessageTypes = append([]MessageType{}, subscription.MessageTypes...)
		}
		client.pluginCfg.WsjtxSubscription = &subCopy
		return nil
	}
}

func WithRawDecodeDelivery() Option {
	return func(client *ClhClient) error {
		if client.pluginCfg.WsjtxSubscription == nil {
			client.pluginCfg.WsjtxSubscription = &WsjtxSubscription{}
		}
		client.pluginCfg.WsjtxSubscription.DecodeDeliveryMode = DecodeDeliveryMode_REALTIME
		return nil
	}
}

func WithWsjtxMessageFilter(types ...MessageType) Option {
	return func(client *ClhClient) error {
		if client.pluginCfg.WsjtxSubscription == nil {
			client.pluginCfg.WsjtxSubscription = &WsjtxSubscription{}
		}
		client.pluginCfg.WsjtxSubscription.MessageTypes = append([]MessageType{}, types...)
		return nil
	}
}

type PluginConfig struct {
	Uuid              string
	Name              string
	Version           string
	Description       string
	Capabilities      []PluginCapability
	Metadata          map[string]string
	SdkName           string
	SdkVersion        string
	WsjtxSubscription *WsjtxSubscription
}

type ClhClient struct {
	conn   net.Conn
	reader *bufio.Reader
	ctx    context.Context
	cancel context.CancelCauseFunc

	wg sync.WaitGroup
	mu sync.Mutex

	closed bool

	pluginCfg         *PluginConfig
	heartbeatInterval time.Duration
	serverInfo        *PipeServerInfo
}

func NewClient(config PluginConfig, opts ...Option) (*ClhClient, error) {
	if config.Uuid == "" {
		return nil, fmt.Errorf("uuid shouldn't be empty")
	}

	if config.Name == "" {
		return nil, fmt.Errorf("name shouldn't be empty")
	}

	if config.Version == "" {
		return nil, fmt.Errorf("version shouldn't be empty")
	}

	if config.Description == "" {
		return nil, fmt.Errorf("description shouldn't be empty")
	}

	if config.Metadata == nil {
		config.Metadata = make(map[string]string)
	}
	if config.SdkName == "" {
		config.SdkName = DefaultSdkName
	}
	if config.SdkVersion == "" {
		config.SdkVersion = DefaultSdkVersion
	}

	c := &ClhClient{
		conn:              nil,
		reader:            nil,
		pluginCfg:         &config,
		heartbeatInterval: DefaultHeartbeatInterval,
	}

	ctx, cancelFunc := context.WithCancelCause(context.Background())
	c.ctx = ctx
	c.cancel = cancelFunc

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("apply option failed: %w", err)
		}
	}

	return c, nil
}

func toPbWsjtxSubscription(subscription *WsjtxSubscription) *plugin.PipeWsjtxSubscription {
	if subscription == nil {
		return nil
	}

	pbSubscription := &plugin.PipeWsjtxSubscription{
		DecodeDeliveryMode: plugin.DecodeDeliveryMode(subscription.DecodeDeliveryMode),
	}

	if len(subscription.MessageTypes) > 0 {
		pbTypes := make([]plugin.MessageType, len(subscription.MessageTypes))
		for i, v := range subscription.MessageTypes {
			pbTypes[i] = plugin.MessageType(v)
		}
		pbSubscription.MessageTypes = pbTypes
	}

	return pbSubscription
}

func (c *ClhClient) Connect() error {
	if c.closed {
		return fmt.Errorf("you are not allowd to call connect on a closed client. please create a new client instead")
	}

	if c.conn != nil {
		return fmt.Errorf("client is already connected")
	}

	conn, err := dialPipe()
	if err != nil {
		return fmt.Errorf("dial pipe failed: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)

	registerReq := &plugin.PipeRegisterPluginReq{
		Uuid:              c.pluginCfg.Uuid,
		Name:              c.pluginCfg.Name,
		Version:           c.pluginCfg.Version,
		Description:       c.pluginCfg.Description,
		Metadata:          c.pluginCfg.Metadata,
		SdkName:           c.pluginCfg.SdkName,
		SdkVersion:        c.pluginCfg.SdkVersion,
		WsjtxSubscription: toPbWsjtxSubscription(c.pluginCfg.WsjtxSubscription),
		Timestamp:         timestamppb.Now(),
	}

	tmp := make([]plugin.Capability, len(c.pluginCfg.Capabilities))
	for i, v := range c.pluginCfg.Capabilities {
		tmp[i] = plugin.Capability(v)
	}
	registerReq.Capabilities = tmp

	_, err = protodelim.MarshalTo(conn, registerReq)
	if err != nil {
		_ = conn.Close()
		c.conn = nil
		c.reader = nil
		return err
	}

	var resp plugin.PipeRegisterPluginResp
	err = protodelim.UnmarshalFrom(c.reader, &resp)
	if err != nil {
		_ = conn.Close()
		c.conn = nil
		c.reader = nil
		return err
	}
	if !resp.Success {
		_ = conn.Close()
		c.conn = nil
		c.reader = nil
		return fmt.Errorf("register plugin failed: %s", resp.Message)
	}

	c.mu.Lock()
	c.serverInfo = convertPipeServerInfo(resp.ServerInfo)
	c.mu.Unlock()

	c.wg.Add(1)
	go c.heartbeat()
	return nil
}

func (c *ClhClient) GetServerInfo() *PipeServerInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.serverInfo == nil {
		return nil
	}

	infoCopy := *c.serverInfo
	return &infoCopy
}

func (c *ClhClient) WaitMessage() (Message, error) {
	anyMsg := &anypb.Any{}
	err := protodelim.UnmarshalFrom(c.reader, anyMsg)
	if err != nil {
		ce := c.ctx.Err()
		if ce != nil {
			return nil, context.Cause(c.ctx)
		}
		return nil, fmt.Errorf("failed to unmarshal: %v", err)
	}

	msg, err := anyMsg.UnmarshalNew()
	if err != nil {
		return nil, err
	}

	switch v := msg.(type) {
	case *plugin.RigData:
		return convertRigData(v), nil
	case *plugin.WsjtxMessage:
		return convertWsjtxMessage(v), nil
	case *plugin.PackedDecodeMessage:
		return convertPackedDecodeMessage(v), nil
	case *plugin.PipeConnectionClosed:
		return &PipeConnectionClosed{Timestamp: v.Timestamp.AsTime()}, nil
	case *plugin.ClhInternalMessage:
		return convertClhInternalMessage(v), nil
	case *plugin.PipeControlResponse:
		return convertPipeControlResponse(v), nil
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
}

func (c *ClhClient) heartbeat() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()
	defer c.wg.Done()

	for {
		select {
		case <-ticker.C:
			hb := &plugin.PipeHeartbeat{
				Uuid:      c.pluginCfg.Uuid,
				Timestamp: timestamppb.Now(),
			}
			_ = c.writeAnyMessageToPipe(hb)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *ClhClient) SendControlRequest(
	requestID string,
	command PipeControlCommand,
	subscription *WsjtxSubscription,
	arguments map[string]string,
) error {
	req := &plugin.PipeControlRequest{
		RequestId:         requestID,
		Command:           plugin.PipeControlCommand(command),
		WsjtxSubscription: toPbWsjtxSubscription(subscription),
		Arguments:         arguments,
		Timestamp:         timestamppb.Now(),
	}

	return c.writeAnyMessageToPipe(req)
}

func (c *ClhClient) RequestServerInfo(requestID string) error {
	return c.SendControlRequest(requestID, PipeControlCommand_GET_SERVER_INFO, nil, nil)
}

func (c *ClhClient) RequestConnectedPlugins(requestID string) error {
	return c.SendControlRequest(requestID, PipeControlCommand_GET_CONNECTED_PLUGINS, nil, nil)
}

func (c *ClhClient) UpdateWsjtxSubscription(requestID string, subscription WsjtxSubscription) error {
	return c.SendControlRequest(requestID, PipeControlCommand_SET_WSJTX_SUBSCRIPTION, &subscription, nil)
}

func (c *ClhClient) SendPluginLog(level PipePluginLogLevel, message string, fields map[string]string) error {
	logReq := &plugin.PipePluginLog{
		Uuid:      c.pluginCfg.Uuid,
		Level:     plugin.PipePluginLogLevel(level),
		Message:   message,
		Fields:    fields,
		Timestamp: timestamppb.Now(),
	}

	return c.writeAnyMessageToPipe(logReq)
}

func (c *ClhClient) Close() error {
	if c.closed {
		return nil
	}

	c.cancel(nil)
	c.wg.Wait()
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("close pipe failed: %w", err)
		}
		c.conn = nil
		c.reader = nil
	}
	return nil
}

func (c *ClhClient) writeRawMessageToPipe(msg protoreflect.ProtoMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is not established")
	}

	_, err := protodelim.MarshalTo(c.conn, msg)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClhClient) writeAnyMessageToPipe(msg protoreflect.ProtoMessage) error {
	anyMsg, err := anypb.New(msg)
	if err != nil {
		return err
	}

	return c.writeRawMessageToPipe(anyMsg)
}
