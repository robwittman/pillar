package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestWebhookEmitter_EmitCreatesDeliveries(t *testing.T) {
	webhookRepo := &mock.WebhookRepository{
		FindByEventTypeFn: func(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
			return []*domain.Webhook{
				{ID: "wh-1"},
				{ID: "wh-2"},
			}, nil
		},
	}

	var mu sync.Mutex
	var deliveries []*domain.WebhookDelivery
	deliveryRepo := &mock.WebhookDeliveryRepository{
		CreateFn: func(ctx context.Context, delivery *domain.WebhookDelivery) error {
			mu.Lock()
			deliveries = append(deliveries, delivery)
			mu.Unlock()
			return nil
		},
	}

	emitter := service.NewWebhookEmitter(webhookRepo, deliveryRepo, testWebhookLogger())

	event := domain.Event{
		ID:        "evt-1",
		Type:      "agent.created",
		Timestamp: time.Now(),
		Data:      map[string]string{"id": "agent-1"},
	}

	emitter.Emit(context.Background(), event)

	// Wait for goroutine
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, deliveries, 2)
	assert.Equal(t, "wh-1", deliveries[0].WebhookID)
	assert.Equal(t, "wh-2", deliveries[1].WebhookID)
	assert.Equal(t, "agent.created", deliveries[0].EventType)
	assert.Equal(t, domain.DeliveryStatusPending, deliveries[0].Status)
}

func TestWebhookEmitter_NoMatchingWebhooks(t *testing.T) {
	webhookRepo := &mock.WebhookRepository{
		FindByEventTypeFn: func(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
			return nil, nil
		},
	}
	deliveryRepo := &mock.WebhookDeliveryRepository{}

	emitter := service.NewWebhookEmitter(webhookRepo, deliveryRepo, testWebhookLogger())

	event := domain.Event{ID: "evt-1", Type: "agent.created", Timestamp: time.Now()}
	emitter.Emit(context.Background(), event)

	time.Sleep(50 * time.Millisecond)
	// No deliveries created - no panic
}

func TestWebhookEmitter_FindWebhooksError(t *testing.T) {
	webhookRepo := &mock.WebhookRepository{
		FindByEventTypeFn: func(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
			return nil, assert.AnError
		},
	}
	deliveryRepo := &mock.WebhookDeliveryRepository{}

	emitter := service.NewWebhookEmitter(webhookRepo, deliveryRepo, testWebhookLogger())

	event := domain.Event{ID: "evt-1", Type: "agent.created", Timestamp: time.Now()}
	emitter.Emit(context.Background(), event)

	time.Sleep(50 * time.Millisecond)
	// No panic on error
}
