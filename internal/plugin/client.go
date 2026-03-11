package plugin

import (
	"context"
	"fmt"
	"net"
	"time"

	pluginv1 "github.com/robwittman/pillar/gen/proto/pillar/plugin/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps a gRPC connection to a plugin process.
type Client struct {
	conn   *grpc.ClientConn
	plugin pluginv1.PluginServiceClient
}

// NewClient dials a plugin over a unix socket.
func NewClient(socketPath string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to plugin at %s: %w", socketPath, err)
	}

	return &Client{
		conn:   conn,
		plugin: pluginv1.NewPluginServiceClient(conn),
	}, nil
}

// Configure sends configuration to the plugin.
func (c *Client) Configure(ctx context.Context, config map[string]string) error {
	resp, err := c.plugin.Configure(ctx, &pluginv1.ConfigureRequest{Config: config})
	if err != nil {
		return fmt.Errorf("configure RPC failed: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("plugin configure failed: %s", resp.Error)
	}
	return nil
}

// OnEvent sends an event to the plugin and returns attribute writes.
func (c *Client) OnEvent(ctx context.Context, req *pluginv1.EventRequest) (*pluginv1.EventResponse, error) {
	return c.plugin.OnEvent(ctx, req)
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// WaitForSocket polls until the socket file exists and is connectable.
func WaitForSocket(socketPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", socketPath, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for plugin socket at %s", socketPath)
}
