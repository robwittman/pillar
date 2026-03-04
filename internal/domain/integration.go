package domain

import (
	"context"
	"time"
)

type Integration struct {
	ID         string         `json:"id"`
	AgentID    string         `json:"agent_id"`
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Config     map[string]any `json:"config"`
	TemplateID string         `json:"template_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type IntegrationRepository interface {
	Create(ctx context.Context, integration *Integration) error
	Get(ctx context.Context, id string) (*Integration, error)
	List(ctx context.Context, agentID string) ([]*Integration, error)
	Update(ctx context.Context, integration *Integration) error
	Delete(ctx context.Context, id string) error
}
