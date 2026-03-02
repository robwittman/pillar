package domain

import (
	"context"
	"time"
)

type AgentStatus string

const (
	AgentStatusPending  AgentStatus = "pending"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusStopped  AgentStatus = "stopped"
	AgentStatusError    AgentStatus = "error"
)

type Agent struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    AgentStatus       `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type AgentRepository interface {
	Create(ctx context.Context, agent *Agent) error
	Get(ctx context.Context, id string) (*Agent, error)
	List(ctx context.Context) ([]*Agent, error)
	Update(ctx context.Context, agent *Agent) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status AgentStatus) error
}

type AgentStatusStore interface {
	SetHeartbeat(ctx context.Context, agentID string, ttl time.Duration) error
	IsOnline(ctx context.Context, agentID string) (bool, error)
	SetOnline(ctx context.Context, agentID string) error
	SetOffline(ctx context.Context, agentID string) error
	ListOnline(ctx context.Context) ([]string, error)
}
