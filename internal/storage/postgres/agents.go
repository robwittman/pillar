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

	err = r.pool.QueryRow(ctx,
		`INSERT INTO agents (id, name, status, metadata, labels)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at, updated_at`,
		agent.ID, agent.Name, agent.Status, metadata, labels,
	).Scan(&agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting agent: %w", err)
	}
	return nil
}

func (r *AgentRepository) Get(ctx context.Context, id string) (*domain.Agent, error) {
	agent := &domain.Agent{}
	var metadata, labels []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, status, metadata, labels, created_at, updated_at
		 FROM agents WHERE id = $1`, id,
	).Scan(&agent.ID, &agent.Name, &agent.Status, &metadata, &labels, &agent.CreatedAt, &agent.UpdatedAt)
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
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, status, metadata, labels, created_at, updated_at
		 FROM agents ORDER BY created_at DESC`)
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

	tag, err := r.pool.Exec(ctx,
		`UPDATE agents SET name = $2, status = $3, metadata = $4, labels = $5
		 WHERE id = $1`,
		agent.ID, agent.Name, agent.Status, metadata, labels,
	)
	if err != nil {
		return fmt.Errorf("updating agent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAgentNotFound
	}
	return nil
}

func (r *AgentRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM agents WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting agent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAgentNotFound
	}
	return nil
}

func (r *AgentRepository) UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE agents SET status = $2 WHERE id = $1`,
		id, status,
	)
	if err != nil {
		return fmt.Errorf("updating agent status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAgentNotFound
	}
	return nil
}
