package service

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

// CompositeEmitter fans out events to multiple emitters in order.
type CompositeEmitter struct {
	emitters []EventEmitter
}

func NewCompositeEmitter(emitters ...EventEmitter) *CompositeEmitter {
	return &CompositeEmitter{emitters: emitters}
}

func (c *CompositeEmitter) Emit(ctx context.Context, event domain.Event) {
	for _, e := range c.emitters {
		e.Emit(ctx, event)
	}
}
