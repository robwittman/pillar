package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type AgentRepository struct {
	CreateFn       func(ctx context.Context, agent *domain.Agent) error
	GetFn          func(ctx context.Context, id string) (*domain.Agent, error)
	ListFn         func(ctx context.Context) ([]*domain.Agent, error)
	UpdateFn       func(ctx context.Context, agent *domain.Agent) error
	DeleteFn       func(ctx context.Context, id string) error
	UpdateStatusFn func(ctx context.Context, id string, status domain.AgentStatus) error
}

func (m *AgentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	return m.CreateFn(ctx, agent)
}

func (m *AgentRepository) Get(ctx context.Context, id string) (*domain.Agent, error) {
	return m.GetFn(ctx, id)
}

func (m *AgentRepository) List(ctx context.Context) ([]*domain.Agent, error) {
	return m.ListFn(ctx)
}

func (m *AgentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	return m.UpdateFn(ctx, agent)
}

func (m *AgentRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *AgentRepository) UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}
