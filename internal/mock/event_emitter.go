package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type EventEmitter struct {
	EmitFn func(ctx context.Context, event domain.Event)
}

func (m *EventEmitter) Emit(ctx context.Context, event domain.Event) {
	if m.EmitFn != nil {
		m.EmitFn(ctx, event)
	}
}
