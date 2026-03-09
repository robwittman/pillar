package domain

import (
	"context"
	"encoding/json"
	"time"
)

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
)

type WebhookDelivery struct {
	ID            string          `json:"id"`
	WebhookID     string          `json:"webhook_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	ResponseCode  int             `json:"response_code,omitempty"`
	ResponseBody  string          `json:"response_body,omitempty"`
	Status        DeliveryStatus  `json:"status"`
	Attempts      int             `json:"attempts"`
	LastAttemptAt *time.Time      `json:"last_attempt_at,omitempty"`
	NextRetryAt   *time.Time      `json:"next_retry_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery *WebhookDelivery) error
	ListPending(ctx context.Context, limit int) ([]*WebhookDelivery, error)
	Update(ctx context.Context, delivery *WebhookDelivery) error
	ListByWebhook(ctx context.Context, webhookID string) ([]*WebhookDelivery, error)
}
