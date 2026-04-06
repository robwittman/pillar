package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type WebhookDeliveryRepository struct {
	CreateFn        func(ctx context.Context, delivery *domain.WebhookDelivery) error
	ListPendingFn   func(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error)
	UpdateFn        func(ctx context.Context, delivery *domain.WebhookDelivery) error
	ListByWebhookFn func(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error)
}

func (m *WebhookDeliveryRepository) Create(ctx context.Context, delivery *domain.WebhookDelivery) error {
	return m.CreateFn(ctx, delivery)
}

func (m *WebhookDeliveryRepository) ListPending(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error) {
	return m.ListPendingFn(ctx, limit)
}

func (m *WebhookDeliveryRepository) Update(ctx context.Context, delivery *domain.WebhookDelivery) error {
	return m.UpdateFn(ctx, delivery)
}

func (m *WebhookDeliveryRepository) ListByWebhook(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error) {
	return m.ListByWebhookFn(ctx, webhookID)
}
