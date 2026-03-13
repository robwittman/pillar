package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type TriggerRepository struct {
	CreateFn       func(ctx context.Context, trigger *domain.Trigger) error
	GetFn          func(ctx context.Context, id string) (*domain.Trigger, error)
	ListFn         func(ctx context.Context) ([]*domain.Trigger, error)
	ListBySourceFn func(ctx context.Context, sourceID string) ([]*domain.Trigger, error)
	UpdateFn       func(ctx context.Context, trigger *domain.Trigger) error
	DeleteFn       func(ctx context.Context, id string) error
}

func (m *TriggerRepository) Create(ctx context.Context, trigger *domain.Trigger) error {
	return m.CreateFn(ctx, trigger)
}

func (m *TriggerRepository) Get(ctx context.Context, id string) (*domain.Trigger, error) {
	return m.GetFn(ctx, id)
}

func (m *TriggerRepository) List(ctx context.Context) ([]*domain.Trigger, error) {
	return m.ListFn(ctx)
}

func (m *TriggerRepository) ListBySource(ctx context.Context, sourceID string) ([]*domain.Trigger, error) {
	return m.ListBySourceFn(ctx, sourceID)
}

func (m *TriggerRepository) Update(ctx context.Context, trigger *domain.Trigger) error {
	return m.UpdateFn(ctx, trigger)
}

func (m *TriggerRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
