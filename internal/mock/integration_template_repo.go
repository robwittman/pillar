package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type IntegrationTemplateRepository struct {
	CreateFn                func(ctx context.Context, template *domain.IntegrationTemplate) error
	GetFn                   func(ctx context.Context, id string) (*domain.IntegrationTemplate, error)
	ListFn                  func(ctx context.Context) ([]*domain.IntegrationTemplate, error)
	UpdateFn                func(ctx context.Context, template *domain.IntegrationTemplate) error
	DeleteFn                func(ctx context.Context, id string) error
	FindMatchingTemplatesFn func(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error)
}

func (m *IntegrationTemplateRepository) Create(ctx context.Context, template *domain.IntegrationTemplate) error {
	return m.CreateFn(ctx, template)
}

func (m *IntegrationTemplateRepository) Get(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
	return m.GetFn(ctx, id)
}

func (m *IntegrationTemplateRepository) List(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
	return m.ListFn(ctx)
}

func (m *IntegrationTemplateRepository) Update(ctx context.Context, template *domain.IntegrationTemplate) error {
	return m.UpdateFn(ctx, template)
}

func (m *IntegrationTemplateRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *IntegrationTemplateRepository) FindMatchingTemplates(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error) {
	return m.FindMatchingTemplatesFn(ctx, labels)
}
