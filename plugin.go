package pluginsdk

import (
	"bufio"
	"context"
	"fmt"
	plugin "github.com/SydneyOwl/clh-proto/gen/go"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"sync"
	"time"
)

const (
	DefaultHeartbeatInterval = time.Second * 5
)

type PluginCapability int32

const (
	CapabilityWsjtxMessage PluginCapability = PluginCapability(plugin.Capability_wsjtx_message)
	CapabilityRigData      PluginCapability = PluginCapability(plugin.Capability_rig_data)
)

type Option func(client *ClhClient) error

func WithHeartbeatInterval(interval time.Duration) Option {
	return func(client *ClhClient) error {
		if interval < time.Second || interval > time.Second*10 {
			return fmt.Errorf("heartbeat interval should btw 1s and 10s")
		}

		client.heartbeatInterval = interval
		return nil
	}
}

type PluginConfig struct {
	Uuid         string
	Name         string
	Version      string
	Description  string
	Capabilities []PluginCapability
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

	c := &ClhClient{
		conn:              nil,
		reader:            nil,
		pluginCfg:         &config,
		heartbeatInterval: DefaultHeartbeatInterval,
	}

	ctx, cancelFunc := context.WithCancelCause(context.Background())
	c.ctx = ctx
	c.cancel = cancelFunc

	if opts != nil {
		for _, opt := range opts {
			if opt == nil {
				continue
			}
			if err := opt(c); err != nil {
				return nil, fmt.Errorf("apply option failed: %w", err)
			}
		}
	}
	return c, nil
}

func (c *ClhClient) Connect() error {
	if c.closed {
		return fmt.Errorf("you are not allowd to call connect on a closed client. please create a new client instead")
	}

	conn, err := dialPipe()
	if err != nil {
		return fmt.Errorf("dial pipe failed: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)

	registerReq := &plugin.PipeRegisterPluginReq{
		Uuid:        c.pluginCfg.Uuid,
		Name:        c.pluginCfg.Name,
		Version:     c.pluginCfg.Version,
		Description: c.pluginCfg.Description,
	}

	tmp := make([]plugin.Capability, len(c.pluginCfg.Capabilities))

	for i, v := range c.pluginCfg.Capabilities {
		tmp[i] = plugin.Capability(v)
	}

	registerReq.Capabilities = tmp

	_, err = protodelim.MarshalTo(conn, registerReq)
	if err != nil {
		return err
	}

	var resp plugin.PipeRegisterPluginResp
	err = protodelim.UnmarshalFrom(c.reader, &resp)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("register plugin failed: %s", resp.Message)
	}

	c.wg.Add(1)
	go c.heartbeat()
	return nil
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
		log.Fatalf("Failed to create message: %v", err)
	}

	switch v := msg.(type) {
	case *plugin.RigData:
		return convertRigData(v), nil
	case *plugin.WsjtxMessage:
		return convertWsjtxMessage(v), nil
	case *plugin.PackedWsjtxMessage:
		return convertPackedWsjtxMessage(v), nil

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
			_ = c.writeMessageToPipe(hb)
		case <-c.ctx.Done():
			return
		}
	}
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
	}
	return nil
}

func (c *ClhClient) writeMessageToPipe(msg protoreflect.ProtoMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := protodelim.MarshalTo(c.conn, msg)

	if err != nil {
		return err
	}
	return nil
}
