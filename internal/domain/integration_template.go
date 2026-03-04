package domain

import (
	"context"
	"time"
)

type IntegrationTemplate struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	Config    map[string]any    `json:"config"`
	Selector  map[string]string `json:"selector"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type IntegrationTemplateRepository interface {
	Create(ctx context.Context, template *IntegrationTemplate) error
	Get(ctx context.Context, id string) (*IntegrationTemplate, error)
	List(ctx context.Context) ([]*IntegrationTemplate, error)
	Update(ctx context.Context, template *IntegrationTemplate) error
	Delete(ctx context.Context, id string) error
	FindMatchingTemplates(ctx context.Context, labels map[string]string) ([]*IntegrationTemplate, error)
}
