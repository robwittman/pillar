package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type IntegrationRepository struct {
	CreateFn func(ctx context.Context, integration *domain.Integration) error
	GetFn    func(ctx context.Context, id string) (*domain.Integration, error)
	ListFn   func(ctx context.Context, agentID string) ([]*domain.Integration, error)
	UpdateFn func(ctx context.Context, integration *domain.Integration) error
	DeleteFn func(ctx context.Context, id string) error
}

func (m *IntegrationRepository) Create(ctx context.Context, integration *domain.Integration) error {
	return m.CreateFn(ctx, integration)
}

func (m *IntegrationRepository) Get(ctx context.Context, id string) (*domain.Integration, error) {
	return m.GetFn(ctx, id)
}

func (m *IntegrationRepository) List(ctx context.Context, agentID string) ([]*domain.Integration, error) {
	return m.ListFn(ctx, agentID)
}

func (m *IntegrationRepository) Update(ctx context.Context, integration *domain.Integration) error {
	return m.UpdateFn(ctx, integration)
}

func (m *IntegrationRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
