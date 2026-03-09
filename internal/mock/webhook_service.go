package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type WebhookService struct {
	CreateFn         func(ctx context.Context, url, description string, eventTypes []string) (*domain.Webhook, string, error)
	GetFn            func(ctx context.Context, id string) (*domain.Webhook, error)
	ListFn           func(ctx context.Context) ([]*domain.Webhook, error)
	UpdateFn         func(ctx context.Context, id string, description string, eventTypes []string, status domain.WebhookStatus) (*domain.Webhook, error)
	DeleteFn         func(ctx context.Context, id string) error
	RotateSecretFn   func(ctx context.Context, id string) (*domain.Webhook, string, error)
	ListDeliveriesFn func(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error)
}

func (m *WebhookService) Create(ctx context.Context, url, description string, eventTypes []string) (*domain.Webhook, string, error) {
	return m.CreateFn(ctx, url, description, eventTypes)
}

func (m *WebhookService) Get(ctx context.Context, id string) (*domain.Webhook, error) {
	return m.GetFn(ctx, id)
}

func (m *WebhookService) List(ctx context.Context) ([]*domain.Webhook, error) {
	return m.ListFn(ctx)
}

func (m *WebhookService) Update(ctx context.Context, id string, description string, eventTypes []string, status domain.WebhookStatus) (*domain.Webhook, error) {
	return m.UpdateFn(ctx, id, description, eventTypes, status)
}

func (m *WebhookService) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *WebhookService) RotateSecret(ctx context.Context, id string) (*domain.Webhook, string, error) {
	return m.RotateSecretFn(ctx, id)
}

func (m *WebhookService) ListDeliveries(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error) {
	return m.ListDeliveriesFn(ctx, webhookID)
}

