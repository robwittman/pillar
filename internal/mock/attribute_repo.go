package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type AgentAttributeRepository struct {
	SetFn               func(ctx context.Context, attr *domain.AgentAttribute) error
	GetFn               func(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error)
	ListFn              func(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error)
	DeleteFn            func(ctx context.Context, agentID, namespace string) error
	DeleteAllForAgentFn func(ctx context.Context, agentID string) error
}

func (m *AgentAttributeRepository) Set(ctx context.Context, attr *domain.AgentAttribute) error {
	return m.SetFn(ctx, attr)
}

func (m *AgentAttributeRepository) Get(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
	return m.GetFn(ctx, agentID, namespace)
}

func (m *AgentAttributeRepository) List(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error) {
	return m.ListFn(ctx, agentID)
}

func (m *AgentAttributeRepository) Delete(ctx context.Context, agentID, namespace string) error {
	return m.DeleteFn(ctx, agentID, namespace)
}

func (m *AgentAttributeRepository) DeleteAllForAgent(ctx context.Context, agentID string) error {
	return m.DeleteAllForAgentFn(ctx, agentID)
}
