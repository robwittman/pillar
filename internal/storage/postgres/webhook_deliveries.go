package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type WebhookDeliveryRepository struct {
	pool *pgxpool.Pool
}

func NewWebhookDeliveryRepository(pool *pgxpool.Pool) *WebhookDeliveryRepository {
	return &WebhookDeliveryRepository{pool: pool}
}

func (r *WebhookDeliveryRepository) Create(ctx context.Context, delivery *domain.WebhookDelivery) error {
	orgID := orgIDFromContext(ctx)
	err := r.pool.QueryRow(ctx,
		`INSERT INTO webhook_deliveries (id, webhook_id, event_type, payload, status, next_retry_at, org_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at`,
		delivery.ID, delivery.WebhookID, delivery.EventType, delivery.Payload,
		delivery.Status, delivery.NextRetryAt, nullIfEmpty(orgID),
	).Scan(&delivery.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting webhook delivery: %w", err)
	}
	return nil
}

func (r *WebhookDeliveryRepository) ListPending(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error) {
	// ListPending is called by the background worker without org context — it processes all pending deliveries.
	rows, err := r.pool.Query(ctx,
		`SELECT id, webhook_id, event_type, payload, response_code, response_body,
		        status, attempts, last_attempt_at, next_retry_at, created_at
		 FROM webhook_deliveries
		 WHERE status = 'pending' AND (next_retry_at IS NULL OR next_retry_at <= now())
		 ORDER BY created_at ASC
		 LIMIT $1`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying pending deliveries: %w", err)
	}
	defer rows.Close()

	return scanDeliveries(rows)
}

func (r *WebhookDeliveryRepository) Update(ctx context.Context, delivery *domain.WebhookDelivery) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE webhook_deliveries
		 SET response_code = $2, response_body = $3, status = $4, attempts = $5,
		     last_attempt_at = $6, next_retry_at = $7
		 WHERE id = $1`,
		delivery.ID, delivery.ResponseCode, delivery.ResponseBody, delivery.Status,
		delivery.Attempts, delivery.LastAttemptAt, delivery.NextRetryAt,
	)
	if err != nil {
		return fmt.Errorf("updating webhook delivery: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delivery not found: %s", delivery.ID)
	}
	return nil
}

func (r *WebhookDeliveryRepository) ListByWebhook(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, webhook_id, event_type, payload, response_code, response_body,
		        status, attempts, last_attempt_at, next_retry_at, created_at
		 FROM webhook_deliveries
		 WHERE webhook_id = $1
		 ORDER BY created_at DESC`, webhookID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying deliveries by webhook: %w", err)
	}
	defer rows.Close()

	return scanDeliveries(rows)
}

func scanDeliveries(rows pgx.Rows) ([]*domain.WebhookDelivery, error) {
	var deliveries []*domain.WebhookDelivery
	for rows.Next() {
		d := &domain.WebhookDelivery{}
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.EventType, &d.Payload,
			&d.ResponseCode, &d.ResponseBody, &d.Status, &d.Attempts,
			&d.LastAttemptAt, &d.NextRetryAt, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning delivery: %w", err)
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}
