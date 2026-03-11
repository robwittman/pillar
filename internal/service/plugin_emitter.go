package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	pluginv1 "github.com/robwittman/pillar/gen/proto/pillar/plugin/v1"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/plugin"
)

// PluginEmitter dispatches events to all registered plugins synchronously
// and writes returned attributes via the attribute service.
type PluginEmitter struct {
	manager *plugin.Manager
	attrSvc AttributeService
	logger  *slog.Logger
}

func NewPluginEmitter(manager *plugin.Manager, attrSvc AttributeService, logger *slog.Logger) *PluginEmitter {
	return &PluginEmitter{
		manager: manager,
		attrSvc: attrSvc,
		logger:  logger,
	}
}

func (e *PluginEmitter) Emit(ctx context.Context, event domain.Event) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		e.logger.Warn("failed to marshal event data for plugins", "event_type", event.Type, "error", err)
		return
	}

	req := &pluginv1.EventRequest{
		Id:        event.ID,
		Type:      event.Type,
		Timestamp: event.Timestamp.Format(time.RFC3339),
		Data:      data,
	}

	for _, p := range e.manager.Plugins() {
		resp, err := p.Client.OnEvent(ctx, req)
		if err != nil {
			e.logger.Warn("plugin OnEvent failed",
				"plugin", p.Name,
				"event_type", event.Type,
				"error", err,
			)
			continue
		}

		if !resp.Success {
			e.logger.Warn("plugin returned error",
				"plugin", p.Name,
				"event_type", event.Type,
				"error", resp.Error,
			)
			continue
		}

		// Process attribute writes
		for _, aw := range resp.Attributes {
			if _, err := e.attrSvc.Set(ctx, aw.AgentId, aw.Namespace, json.RawMessage(aw.Value)); err != nil {
				e.logger.Warn("failed to write plugin attribute",
					"plugin", p.Name,
					"agent_id", aw.AgentId,
					"namespace", aw.Namespace,
					"error", err,
				)
			}
		}
	}
}
