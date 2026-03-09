package domain

import (
	"context"
	"encoding/json"
	"time"
)

type AgentAttribute struct {
	AgentID   string          `json:"agent_id"`
	Namespace string          `json:"namespace"`
	Value     json.RawMessage `json:"value"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type AgentAttributeRepository interface {
	Set(ctx context.Context, attr *AgentAttribute) error
	Get(ctx context.Context, agentID, namespace string) (*AgentAttribute, error)
	List(ctx context.Context, agentID string) ([]*AgentAttribute, error)
	Delete(ctx context.Context, agentID, namespace string) error
	DeleteAllForAgent(ctx context.Context, agentID string) error
}
