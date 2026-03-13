package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type TriggerService struct {
	CreateFn       func(ctx context.Context, sourceID, agentID, name string, filter domain.TriggerFilter, taskTemplate string) (*domain.Trigger, error)
	GetFn          func(ctx context.Context, id string) (*domain.Trigger, error)
	ListFn         func(ctx context.Context) ([]*domain.Trigger, error)
	ListBySourceFn func(ctx context.Context, sourceID string) ([]*domain.Trigger, error)
	UpdateFn       func(ctx context.Context, id string, name string, filter *domain.TriggerFilter, taskTemplate *string, enabled *bool) (*domain.Trigger, error)
	DeleteFn       func(ctx context.Context, id string) error
}

func (m *TriggerService) Create(ctx context.Context, sourceID, agentID, name string, filter domain.TriggerFilter, taskTemplate string) (*domain.Trigger, error) {
	return m.CreateFn(ctx, sourceID, agentID, name, filter, taskTemplate)
}

func (m *TriggerService) Get(ctx context.Context, id string) (*domain.Trigger, error) {
	return m.GetFn(ctx, id)
}

func (m *TriggerService) List(ctx context.Context) ([]*domain.Trigger, error) {
	return m.ListFn(ctx)
}

func (m *TriggerService) ListBySource(ctx context.Context, sourceID string) ([]*domain.Trigger, error) {
	return m.ListBySourceFn(ctx, sourceID)
}

func (m *TriggerService) Update(ctx context.Context, id string, name string, filter *domain.TriggerFilter, taskTemplate *string, enabled *bool) (*domain.Trigger, error) {
	return m.UpdateFn(ctx, id, name, filter, taskTemplate, enabled)
}

func (m *TriggerService) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
