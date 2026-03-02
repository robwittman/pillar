package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type AgentConfigRepository struct {
	CreateFn func(ctx context.Context, config *domain.AgentConfig) error
	GetFn    func(ctx context.Context, agentID string) (*domain.AgentConfig, error)
	UpdateFn func(ctx context.Context, config *domain.AgentConfig) error
	DeleteFn func(ctx context.Context, agentID string) error
}

func (m *AgentConfigRepository) Create(ctx context.Context, config *domain.AgentConfig) error {
	return m.CreateFn(ctx, config)
}

func (m *AgentConfigRepository) Get(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
	return m.GetFn(ctx, agentID)
}

func (m *AgentConfigRepository) Update(ctx context.Context, config *domain.AgentConfig) error {
	return m.UpdateFn(ctx, config)
}

func (m *AgentConfigRepository) Delete(ctx context.Context, agentID string) error {
	return m.DeleteFn(ctx, agentID)
}
