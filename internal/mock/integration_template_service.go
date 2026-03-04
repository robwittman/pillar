package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type IntegrationTemplateService struct {
	CreateFn            func(ctx context.Context, template *domain.IntegrationTemplate) error
	GetFn               func(ctx context.Context, id string) (*domain.IntegrationTemplate, error)
	ListFn              func(ctx context.Context) ([]*domain.IntegrationTemplate, error)
	UpdateFn            func(ctx context.Context, template *domain.IntegrationTemplate) error
	DeleteFn            func(ctx context.Context, id string) error
	PreviewFn           func(ctx context.Context, id string) ([]*domain.Agent, error)
	ProvisionForAgentFn func(ctx context.Context, agentID string, labels map[string]string) error
}

func (m *IntegrationTemplateService) Create(ctx context.Context, template *domain.IntegrationTemplate) error {
	return m.CreateFn(ctx, template)
}

func (m *IntegrationTemplateService) Get(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
	return m.GetFn(ctx, id)
}

func (m *IntegrationTemplateService) List(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
	return m.ListFn(ctx)
}

func (m *IntegrationTemplateService) Update(ctx context.Context, template *domain.IntegrationTemplate) error {
	return m.UpdateFn(ctx, template)
}

func (m *IntegrationTemplateService) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *IntegrationTemplateService) Preview(ctx context.Context, id string) ([]*domain.Agent, error) {
	return m.PreviewFn(ctx, id)
}

func (m *IntegrationTemplateService) ProvisionForAgent(ctx context.Context, agentID string, labels map[string]string) error {
	return m.ProvisionForAgentFn(ctx, agentID, labels)
}
