package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type AgentAttributeRepository struct {
	pool *pgxpool.Pool
}

func NewAgentAttributeRepository(pool *pgxpool.Pool) *AgentAttributeRepository {
	return &AgentAttributeRepository{pool: pool}
}

func (r *AgentAttributeRepository) Set(ctx context.Context, attr *domain.AgentAttribute) error {
	orgID := orgIDFromContext(ctx)
	err := r.pool.QueryRow(ctx,
		`INSERT INTO agent_attributes (agent_id, namespace, value, org_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (agent_id, namespace) DO UPDATE SET value = EXCLUDED.value
		 RETURNING created_at, updated_at`,
		attr.AgentID, attr.Namespace, attr.Value, nullIfEmpty(orgID),
	).Scan(&attr.CreatedAt, &attr.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting agent attribute: %w", err)
	}
	return nil
}

func (r *AgentAttributeRepository) Get(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
	attr := &domain.AgentAttribute{}
	orgID := orgIDFromContext(ctx)
	query := `SELECT agent_id, namespace, value, created_at, updated_at
		 FROM agent_attributes WHERE agent_id = $1 AND namespace = $2`
	args := []any{agentID, namespace}
	if orgID != "" {
		query += ` AND org_id = $3`
		args = append(args, orgID)
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&attr.AgentID, &attr.Namespace, &attr.Value, &attr.CreatedAt, &attr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAttributeNotFound
		}
		return nil, fmt.Errorf("querying agent attribute: %w", err)
	}
	return attr, nil
}

func (r *AgentAttributeRepository) List(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT agent_id, namespace, value, created_at, updated_at
		 FROM agent_attributes WHERE agent_id = $1 ORDER BY namespace`, agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying agent attributes: %w", err)
	}
	defer rows.Close()

	var attrs []*domain.AgentAttribute
	for rows.Next() {
		attr := &domain.AgentAttribute{}
		if err := rows.Scan(&attr.AgentID, &attr.Namespace, &attr.Value, &attr.CreatedAt, &attr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning agent attribute: %w", err)
		}
		attrs = append(attrs, attr)
	}
	return attrs, rows.Err()
}

func (r *AgentAttributeRepository) Delete(ctx context.Context, agentID, namespace string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM agent_attributes WHERE agent_id = $1 AND namespace = $2`,
		agentID, namespace,
	)
	if err != nil {
		return fmt.Errorf("deleting agent attribute: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAttributeNotFound
	}
	return nil
}

func (r *AgentAttributeRepository) DeleteAllForAgent(ctx context.Context, agentID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM agent_attributes WHERE agent_id = $1`, agentID,
	)
	if err != nil {
		return fmt.Errorf("deleting all agent attributes: %w", err)
	}
	return nil
}
