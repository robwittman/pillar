package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/plugin"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pluginTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCompositeEmitter_CallsAllEmitters(t *testing.T) {
	var calls []string

	e1 := &mock.EventEmitter{
		EmitFn: func(ctx context.Context, event domain.Event) {
			calls = append(calls, "e1")
		},
	}
	e2 := &mock.EventEmitter{
		EmitFn: func(ctx context.Context, event domain.Event) {
			calls = append(calls, "e2")
		},
	}

	composite := service.NewCompositeEmitter(e1, e2)
	composite.Emit(context.Background(), domain.Event{
		ID:        "test",
		Type:      "agent.created",
		Timestamp: time.Now(),
	})

	require.Len(t, calls, 2)
	assert.Equal(t, "e1", calls[0])
	assert.Equal(t, "e2", calls[1])
}

func TestCompositeEmitter_EmptyEmitters(t *testing.T) {
	composite := service.NewCompositeEmitter()
	// Should not panic
	composite.Emit(context.Background(), domain.Event{
		ID:   "test",
		Type: "agent.created",
	})
}

func TestPluginEmitter_NoPlugins(t *testing.T) {
	mgr := plugin.NewManager(pluginTestLogger())
	attrSvc := &mock.AttributeService{}

	emitter := service.NewPluginEmitter(mgr, attrSvc, pluginTestLogger())
	// Should not panic with no plugins
	emitter.Emit(context.Background(), domain.Event{
		ID:        "test",
		Type:      "agent.created",
		Timestamp: time.Now(),
		Data:      map[string]string{"id": "agent-1"},
	})
}
