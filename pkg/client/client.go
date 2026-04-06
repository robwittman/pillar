package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
)

// DirectiveHandler is called when the server sends a directive to the agent.
type DirectiveHandler func(directiveType, payload string)

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithToken sets the Bearer token used for gRPC authentication.
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}

type Client struct {
	conn   *grpc.ClientConn
	stream pillarv1.AgentStreamService_AgentStreamClient
	logger *slog.Logger

	agentID           string
	token             string
	heartbeatInterval time.Duration
	stopHeartbeat     chan struct{}
	config            *pillarv1.AgentConfig
	attributes        map[string][]byte

	mu          sync.RWMutex
	status      pillarv1.AgentStatus
	onDirective DirectiveHandler
}

func New(addr, agentID string, logger *slog.Logger, opts ...ClientOption) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connecting to pillar: %w", err)
	}

	c := &Client{
		conn:          conn,
		agentID:       agentID,
		logger:        logger,
		stopHeartbeat: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *Client) Connect(ctx context.Context, capabilities map[string]string) error {
	// Attach Bearer token as gRPC metadata if configured.
	if c.token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
	}

	stub := pillarv1.NewAgentStreamServiceClient(c.conn)
	stream, err := stub.AgentStream(ctx)
	if err != nil {
		return fmt.Errorf("opening stream: %w", err)
	}
	c.stream = stream

	if err := stream.Send(&pillarv1.AgentMessage{
		Payload: &pillarv1.AgentMessage_Connect{
			Connect: &pillarv1.ConnectRequest{
				AgentId:      c.agentID,
				Capabilities: capabilities,
			},
		},
	}); err != nil {
		return fmt.Errorf("sending connect: %w", err)
	}

	msg, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receiving connect ack: %w", err)
	}

	ack := msg.GetConnectAck()
	if ack == nil || !ack.Accepted {
		return fmt.Errorf("connection rejected")
	}

	c.heartbeatInterval = time.Duration(ack.HeartbeatIntervalSeconds) * time.Second
	c.config = ack.Config
	c.attributes = ack.Attributes
	c.mu.Lock()
	c.status = ack.Status
	c.mu.Unlock()
	c.logger.Info("connected to pillar", "interval", c.heartbeatInterval, "status", ack.Status, "has_config", c.config != nil)

	go c.heartbeatLoop()
	return nil
}

func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.stream.Send(&pillarv1.AgentMessage{
				Payload: &pillarv1.AgentMessage_Heartbeat{
					Heartbeat: &pillarv1.Heartbeat{
						AgentId:   c.agentID,
						Timestamp: timestamppb.Now(),
					},
				},
			}); err != nil {
				c.logger.Error("heartbeat send failed", "error", err)
				return
			}
		case <-c.stopHeartbeat:
			return
		}
	}
}

func (c *Client) SendEvent(eventType, payload string) error {
	return c.stream.Send(&pillarv1.AgentMessage{
		Payload: &pillarv1.AgentMessage_Event{
			Event: &pillarv1.EventReport{
				AgentId:   c.agentID,
				EventType: eventType,
				Payload:   payload,
				Timestamp: timestamppb.Now(),
			},
		},
	})
}

func (c *Client) SendTaskResult(taskID string, success bool, output, errMsg string) error {
	return c.stream.Send(&pillarv1.AgentMessage{
		Payload: &pillarv1.AgentMessage_TaskResult{
			TaskResult: &pillarv1.TaskResult{
				TaskId:  taskID,
				AgentId: c.agentID,
				Success: success,
				Output:  output,
				Error:   errMsg,
			},
		},
	})
}

func (c *Client) Config() *pillarv1.AgentConfig {
	return c.config
}

// Attributes returns the raw attributes map received at connect time.
func (c *Client) Attributes() map[string][]byte {
	return c.attributes
}

// Attribute returns a single attribute by namespace, or nil if not found.
func (c *Client) Attribute(namespace string) []byte {
	if c.attributes == nil {
		return nil
	}
	return c.attributes[namespace]
}

// Status returns the agent's last-known status from the server.
func (c *Client) Status() pillarv1.AgentStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// OnDirective registers a handler that will be called when the server pushes
// a directive (e.g. start, stop) to this agent. Must be called before Listen().
func (c *Client) OnDirective(handler DirectiveHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onDirective = handler
}

// Listen blocks and dispatches incoming server messages. It handles:
//   - HeartbeatAck: debug log
//   - Directive: updates local status for start/stop, calls onDirective callback
//   - TaskAssignment: log (future iteration)
//
// Returns on EOF or stream error.
func (c *Client) Listen() error {
	for {
		msg, err := c.stream.Recv()
		if err == io.EOF {
			c.logger.Info("server closed stream")
			return nil
		}
		if err != nil {
			return fmt.Errorf("listen recv: %w", err)
		}

		switch p := msg.Payload.(type) {
		case *pillarv1.ServerMessage_HeartbeatAck:
			c.logger.Debug("heartbeat ack", "server_time", p.HeartbeatAck.ServerTime)

		case *pillarv1.ServerMessage_Directive:
			directive := p.Directive
			c.logger.Info("received directive", "type", directive.Type, "payload", directive.Payload)

			switch directive.Type {
			case "start":
				c.mu.Lock()
				c.status = pillarv1.AgentStatus_AGENT_STATUS_RUNNING
				c.mu.Unlock()
			case "stop":
				c.mu.Lock()
				c.status = pillarv1.AgentStatus_AGENT_STATUS_STOPPED
				c.mu.Unlock()
			}

			c.mu.RLock()
			handler := c.onDirective
			c.mu.RUnlock()
			if handler != nil {
				handler(directive.Type, directive.Payload)
			}

		case *pillarv1.ServerMessage_TaskAssignment:
			c.logger.Info("received task assignment", "task_id", p.TaskAssignment.TaskId)

		case *pillarv1.ServerMessage_ConnectAck:
			c.logger.Debug("unexpected connect ack during listen")
		}
	}
}

func (c *Client) Recv() (*pillarv1.ServerMessage, error) {
	return c.stream.Recv()
}

func (c *Client) Close() error {
	close(c.stopHeartbeat)
	if c.stream != nil {
		_ = c.stream.CloseSend()
	}
	return c.conn.Close()
}
