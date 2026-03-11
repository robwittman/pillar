// Package plugin provides an SDK for building Pillar plugins.
//
// A plugin is a standalone binary that serves a gRPC PluginService over a unix socket.
// Pillar starts the plugin binary and communicates with it to dispatch lifecycle events.
//
// Usage:
//
//	func main() {
//	    plugin.Serve(&MyPlugin{})
//	}
package plugin

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pluginv1 "github.com/robwittman/pillar/gen/proto/pillar/plugin/v1"
	"google.golang.org/grpc"
)

// Plugin is the interface that plugin authors implement.
type Plugin interface {
	// Configure is called once after the plugin starts with config from pillar.yaml.
	Configure(config map[string]string) error

	// OnEvent is called for each lifecycle event. Return attribute writes to
	// persist data back to agents.
	OnEvent(event *pluginv1.EventRequest) (*pluginv1.EventResponse, error)
}

// Serve starts a gRPC server on the unix socket specified by PILLAR_PLUGIN_SOCKET
// and blocks until the process receives SIGINT or SIGTERM.
func Serve(p Plugin) {
	socketPath := os.Getenv("PILLAR_PLUGIN_SOCKET")
	if socketPath == "" {
		log.Fatal("PILLAR_PLUGIN_SOCKET not set")
	}

	// Clean up stale socket
	os.Remove(socketPath)

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", socketPath, err)
	}

	srv := grpc.NewServer()
	pluginv1.RegisterPluginServiceServer(srv, &adapter{plugin: p})

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		srv.GracefulStop()
	}()

	fmt.Fprintf(os.Stderr, "plugin listening on %s\n", socketPath)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("plugin serve error: %v", err)
	}
}
