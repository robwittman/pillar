package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

type WebhookEmitter struct {
	webhookRepo  domain.WebhookRepository
	deliveryRepo domain.WebhookDeliveryRepository
	logger       *slog.Logger
}

func NewWebhookEmitter(webhookRepo domain.WebhookRepository, deliveryRepo domain.WebhookDeliveryRepository, logger *slog.Logger) *WebhookEmitter {
	return &WebhookEmitter{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		logger:       logger,
	}
}

func (e *WebhookEmitter) Emit(ctx context.Context, event domain.Event) {
	go e.emit(ctx, event)
}

func (e *WebhookEmitter) emit(ctx context.Context, event domain.Event) {
	webhooks, err := e.webhookRepo.FindByEventType(ctx, event.Type)
	if err != nil {
		e.logger.Warn("failed to find webhooks for event", "event_type", event.Type, "error", err)
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		e.logger.Warn("failed to marshal event", "event_type", event.Type, "error", err)
		return
	}

	for _, webhook := range webhooks {
		delivery := &domain.WebhookDelivery{
			ID:        uuid.New().String(),
			WebhookID: webhook.ID,
			EventType: event.Type,
			Payload:   payload,
			Status:    domain.DeliveryStatusPending,
		}

		if err := e.deliveryRepo.Create(ctx, delivery); err != nil {
			e.logger.Warn("failed to create webhook delivery",
				"webhook_id", webhook.ID,
				"event_type", event.Type,
				"error", err,
			)
		}
	}
}
