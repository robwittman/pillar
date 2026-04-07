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

type TriggerRepository struct {
	pool *pgxpool.Pool
}

func NewTriggerRepository(pool *pgxpool.Pool) *TriggerRepository {
	return &TriggerRepository{pool: pool}
}

func (r *TriggerRepository) Create(ctx context.Context, trigger *domain.Trigger) error {
	filterJSON, err := json.Marshal(trigger.Filter)
	if err != nil {
		return fmt.Errorf("marshaling filter: %w", err)
	}

	orgID := orgIDFromContext(ctx)
	err = r.pool.QueryRow(ctx,
		`INSERT INTO triggers (id, source_id, agent_id, name, filter, task_template, enabled, org_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING created_at, updated_at`,
		trigger.ID, trigger.SourceID, trigger.AgentID, trigger.Name,
		filterJSON, trigger.TaskTemplate, trigger.Enabled, nullIfEmpty(orgID),
	).Scan(&trigger.CreatedAt, &trigger.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting trigger: %w", err)
	}
	return nil
}

func (r *TriggerRepository) Get(ctx context.Context, id string) (*domain.Trigger, error) {
	t := &domain.Trigger{}
	var filterJSON []byte

	orgID := orgIDFromContext(ctx)
	query := `SELECT id, source_id, agent_id, name, filter, task_template, enabled, created_at, updated_at
		 FROM triggers WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.SourceID, &t.AgentID, &t.Name, &filterJSON,
		&t.TaskTemplate, &t.Enabled, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTriggerNotFound
		}
		return nil, fmt.Errorf("querying trigger: %w", err)
	}

	if err := json.Unmarshal(filterJSON, &t.Filter); err != nil {
		return nil, fmt.Errorf("unmarshaling filter: %w", err)
	}
	return t, nil
}

func (r *TriggerRepository) List(ctx context.Context) ([]*domain.Trigger, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, source_id, agent_id, name, filter, task_template, enabled, created_at, updated_at FROM triggers`
	var args []any
	if orgID != "" {
		query += ` WHERE org_id = $1`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying triggers: %w", err)
	}
	defer rows.Close()

	return scanTriggers(rows)
}

func (r *TriggerRepository) ListBySource(ctx context.Context, sourceID string) ([]*domain.Trigger, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, source_id, agent_id, name, filter, task_template, enabled, created_at, updated_at
		 FROM triggers WHERE source_id = $1 AND enabled = true`
	args := []any{sourceID}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying triggers by source: %w", err)
	}
	defer rows.Close()

	return scanTriggers(rows)
}

func (r *TriggerRepository) Update(ctx context.Context, trigger *domain.Trigger) error {
	filterJSON, err := json.Marshal(trigger.Filter)
	if err != nil {
		return fmt.Errorf("marshaling filter: %w", err)
	}

	orgID := orgIDFromContext(ctx)
	query := `UPDATE triggers SET source_id = $2, agent_id = $3, name = $4, filter = $5, task_template = $6, enabled = $7
		 WHERE id = $1`
	args := []any{trigger.ID, trigger.SourceID, trigger.AgentID, trigger.Name,
		filterJSON, trigger.TaskTemplate, trigger.Enabled}
	if orgID != "" {
		query += ` AND org_id = $8`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating trigger: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTriggerNotFound
	}
	return nil
}

func (r *TriggerRepository) Delete(ctx context.Context, id string) error {
	orgID := orgIDFromContext(ctx)
	query := `DELETE FROM triggers WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting trigger: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTriggerNotFound
	}
	return nil
}

func scanTriggers(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*domain.Trigger, error) {
	var triggers []*domain.Trigger
	for rows.Next() {
		t := &domain.Trigger{}
		var filterJSON []byte
		if err := rows.Scan(&t.ID, &t.SourceID, &t.AgentID, &t.Name, &filterJSON, &t.TaskTemplate, &t.Enabled, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning trigger: %w", err)
		}
		if err := json.Unmarshal(filterJSON, &t.Filter); err != nil {
			return nil, fmt.Errorf("unmarshaling filter: %w", err)
		}
		triggers = append(triggers, t)
	}
	return triggers, rows.Err()
}
