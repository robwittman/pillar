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

type AgentRepository struct {
	pool *pgxpool.Pool
}

func NewAgentRepository(pool *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{pool: pool}
}

func (r *AgentRepository) Create(ctx context.Context, agent *domain.Agent) error {
	metadata, err := json.Marshal(agent.Metadata)
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}
	labels, err := json.Marshal(agent.Labels)
	if err != nil {
		return fmt.Errorf("marshaling labels: %w", err)
	}

	orgID := orgIDFromContext(ctx)
	err = r.pool.QueryRow(ctx,
		`INSERT INTO agents (id, name, status, metadata, labels, org_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at, updated_at`,
		agent.ID, agent.Name, agent.Status, metadata, labels, nullIfEmpty(orgID),
	).Scan(&agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting agent: %w", err)
	}
	return nil
}

func (r *AgentRepository) Get(ctx context.Context, id string) (*domain.Agent, error) {
	agent := &domain.Agent{}
	var metadata, labels []byte

	orgID := orgIDFromContext(ctx)
	query := `SELECT id, name, status, metadata, labels, created_at, updated_at FROM agents WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&agent.ID, &agent.Name, &agent.Status, &metadata, &labels, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAgentNotFound
		}
		return nil, fmt.Errorf("querying agent: %w", err)
	}

	if err := json.Unmarshal(metadata, &agent.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshaling metadata: %w", err)
	}
	if err := json.Unmarshal(labels, &agent.Labels); err != nil {
		return nil, fmt.Errorf("unmarshaling labels: %w", err)
	}
	return agent, nil
}

func (r *AgentRepository) List(ctx context.Context) ([]*domain.Agent, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, name, status, metadata, labels, created_at, updated_at FROM agents`
	var args []any
	if orgID != "" {
		query += ` WHERE org_id = $1`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying agents: %w", err)
	}
	defer rows.Close()

	var agents []*domain.Agent
	for rows.Next() {
		agent := &domain.Agent{}
		var metadata, labels []byte
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Status, &metadata, &labels, &agent.CreatedAt, &agent.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning agent: %w", err)
		}
		if err := json.Unmarshal(metadata, &agent.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshaling metadata: %w", err)
		}
		if err := json.Unmarshal(labels, &agent.Labels); err != nil {
			return nil, fmt.Errorf("unmarshaling labels: %w", err)
		}
		agents = append(agents, agent)
	}
	return agents, rows.Err()
}

func (r *AgentRepository) Update(ctx context.Context, agent *domain.Agent) error {
	metadata, err := json.Marshal(agent.Metadata)
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}
	labels, err := json.Marshal(agent.Labels)
	if err != nil {
		return fmt.Errorf("marshaling labels: %w", err)
	}

	orgID := orgIDFromContext(ctx)
	query := `UPDATE agents SET name = $2, status = $3, metadata = $4, labels = $5 WHERE id = $1`
	args := []any{agent.ID, agent.Name, agent.Status, metadata, labels}
	if orgID != "" {
		query += ` AND org_id = $6`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating agent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAgentNotFound
	}
	return nil
}

func (r *AgentRepository) Delete(ctx context.Context, id string) error {
	orgID := orgIDFromContext(ctx)
	query := `DELETE FROM agents WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting agent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAgentNotFound
	}
	return nil
}

func (r *AgentRepository) UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error {
	orgID := orgIDFromContext(ctx)
	query := `UPDATE agents SET status = $2 WHERE id = $1`
	args := []any{id, status}
	if orgID != "" {
		query += ` AND org_id = $3`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating agent status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAgentNotFound
	}
	return nil
}

// nullIfEmpty returns nil for empty strings, allowing nullable org_id columns.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
