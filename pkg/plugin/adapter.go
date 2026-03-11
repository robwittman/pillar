package plugin

import (
	"context"

	pluginv1 "github.com/robwittman/pillar/gen/proto/pillar/plugin/v1"
)

// adapter bridges the Plugin interface to the gRPC PluginServiceServer.
type adapter struct {
	pluginv1.UnimplementedPluginServiceServer
	plugin Plugin
}

func (a *adapter) Configure(_ context.Context, req *pluginv1.ConfigureRequest) (*pluginv1.ConfigureResponse, error) {
	if err := a.plugin.Configure(req.Config); err != nil {
		return &pluginv1.ConfigureResponse{Success: false, Error: err.Error()}, nil
	}
	return &pluginv1.ConfigureResponse{Success: true}, nil
}

func (a *adapter) OnEvent(_ context.Context, req *pluginv1.EventRequest) (*pluginv1.EventResponse, error) {
	return a.plugin.OnEvent(req)
}
