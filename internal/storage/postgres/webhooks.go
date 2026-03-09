package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type WebhookRepository struct {
	pool *pgxpool.Pool
}

func NewWebhookRepository(pool *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{pool: pool}
}

func (r *WebhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	eventTypes, err := json.Marshal(webhook.EventTypes)
	if err != nil {
		return fmt.Errorf("marshaling event_types: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO webhooks (id, url, secret, event_types, status, description)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at, updated_at`,
		webhook.ID, webhook.URL, webhook.Secret, eventTypes, webhook.Status, webhook.Description,
	).Scan(&webhook.CreatedAt, &webhook.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting webhook: %w", err)
	}
	return nil
}

func (r *WebhookRepository) Get(ctx context.Context, id string) (*domain.Webhook, error) {
	webhook := &domain.Webhook{}
	var eventTypes []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, url, secret, event_types, status, description, created_at, updated_at
		 FROM webhooks WHERE id = $1`, id,
	).Scan(&webhook.ID, &webhook.URL, &webhook.Secret, &eventTypes, &webhook.Status, &webhook.Description, &webhook.CreatedAt, &webhook.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWebhookNotFound
		}
		return nil, fmt.Errorf("querying webhook: %w", err)
	}

	if err := json.Unmarshal(eventTypes, &webhook.EventTypes); err != nil {
		return nil, fmt.Errorf("unmarshaling event_types: %w", err)
	}
	return webhook, nil
}

func (r *WebhookRepository) List(ctx context.Context) ([]*domain.Webhook, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, url, secret, event_types, status, description, created_at, updated_at
		 FROM webhooks ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		webhook := &domain.Webhook{}
		var eventTypes []byte
		if err := rows.Scan(&webhook.ID, &webhook.URL, &webhook.Secret, &eventTypes, &webhook.Status, &webhook.Description, &webhook.CreatedAt, &webhook.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		if err := json.Unmarshal(eventTypes, &webhook.EventTypes); err != nil {
			return nil, fmt.Errorf("unmarshaling event_types: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}
	return webhooks, rows.Err()
}

func (r *WebhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	eventTypes, err := json.Marshal(webhook.EventTypes)
	if err != nil {
		return fmt.Errorf("marshaling event_types: %w", err)
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE webhooks SET url = $2, secret = $3, event_types = $4, status = $5, description = $6
		 WHERE id = $1`,
		webhook.ID, webhook.URL, webhook.Secret, eventTypes, webhook.Status, webhook.Description,
	)
	if err != nil {
		return fmt.Errorf("updating webhook: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWebhookNotFound
	}
	return nil
}

func (r *WebhookRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM webhooks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting webhook: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWebhookNotFound
	}
	return nil
}

func (r *WebhookRepository) FindByEventType(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
	filter, err := json.Marshal([]string{eventType})
	if err != nil {
		return nil, fmt.Errorf("marshaling filter: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, url, secret, event_types, status, description, created_at, updated_at
		 FROM webhooks WHERE status = 'active' AND event_types @> $1::jsonb`, filter,
	)
	if err != nil {
		return nil, fmt.Errorf("querying webhooks by event type: %w", err)
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		webhook := &domain.Webhook{}
		var eventTypes []byte
		if err := rows.Scan(&webhook.ID, &webhook.URL, &webhook.Secret, &eventTypes, &webhook.Status, &webhook.Description, &webhook.CreatedAt, &webhook.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		if err := json.Unmarshal(eventTypes, &webhook.EventTypes); err != nil {
			return nil, fmt.Errorf("unmarshaling event_types: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}
	return webhooks, rows.Err()
}
