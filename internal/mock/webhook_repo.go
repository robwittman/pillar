package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type WebhookRepository struct {
	CreateFn          func(ctx context.Context, webhook *domain.Webhook) error
	GetFn             func(ctx context.Context, id string) (*domain.Webhook, error)
	ListFn            func(ctx context.Context) ([]*domain.Webhook, error)
	UpdateFn          func(ctx context.Context, webhook *domain.Webhook) error
	DeleteFn          func(ctx context.Context, id string) error
	FindByEventTypeFn func(ctx context.Context, eventType string) ([]*domain.Webhook, error)
}

func (m *WebhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	return m.CreateFn(ctx, webhook)
}

func (m *WebhookRepository) Get(ctx context.Context, id string) (*domain.Webhook, error) {
	return m.GetFn(ctx, id)
}

func (m *WebhookRepository) List(ctx context.Context) ([]*domain.Webhook, error) {
	return m.ListFn(ctx)
}

func (m *WebhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	return m.UpdateFn(ctx, webhook)
}

func (m *WebhookRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *WebhookRepository) FindByEventType(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
	return m.FindByEventTypeFn(ctx, eventType)
}
