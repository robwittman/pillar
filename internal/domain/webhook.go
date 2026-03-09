package domain

import (
	"context"
	"time"
)

type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
)

type Webhook struct {
	ID          string        `json:"id"`
	URL         string        `json:"url"`
	Secret      string        `json:"-"`
	EventTypes  []string      `json:"event_types"`
	Status      WebhookStatus `json:"status"`
	Description string        `json:"description"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type WebhookRepository interface {
	Create(ctx context.Context, webhook *Webhook) error
	Get(ctx context.Context, id string) (*Webhook, error)
	List(ctx context.Context) ([]*Webhook, error)
	Update(ctx context.Context, webhook *Webhook) error
	Delete(ctx context.Context, id string) error
	FindByEventType(ctx context.Context, eventType string) ([]*Webhook, error)
}
