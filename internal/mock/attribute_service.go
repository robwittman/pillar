package mock

import (
	"context"
	"encoding/json"

	"github.com/robwittman/pillar/internal/domain"
)

type AttributeService struct {
	SetFn    func(ctx context.Context, agentID, namespace string, value json.RawMessage) (*domain.AgentAttribute, error)
	GetFn    func(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error)
	ListFn   func(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error)
	DeleteFn func(ctx context.Context, agentID, namespace string) error
}

func (m *AttributeService) Set(ctx context.Context, agentID, namespace string, value json.RawMessage) (*domain.AgentAttribute, error) {
	return m.SetFn(ctx, agentID, namespace, value)
}

func (m *AttributeService) Get(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
	return m.GetFn(ctx, agentID, namespace)
}

func (m *AttributeService) List(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error) {
	return m.ListFn(ctx, agentID)
}

func (m *AttributeService) Delete(ctx context.Context, agentID, namespace string) error {
	return m.DeleteFn(ctx, agentID, namespace)
}
