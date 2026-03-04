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

type IntegrationRepository struct {
	pool *pgxpool.Pool
}

func NewIntegrationRepository(pool *pgxpool.Pool) *IntegrationRepository {
	return &IntegrationRepository{pool: pool}
}

func (r *IntegrationRepository) Create(ctx context.Context, integration *domain.Integration) error {
	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO integrations (id, agent_id, type, name, config, template_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING created_at, updated_at`,
		integration.ID, integration.AgentID, integration.Type, integration.Name, configJSON, integration.TemplateID,
	).Scan(&integration.CreatedAt, &integration.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrInvalidIntegration
		}
		return fmt.Errorf("inserting integration: %w", err)
	}
	return nil
}

func (r *IntegrationRepository) Get(ctx context.Context, id string) (*domain.Integration, error) {
	integration := &domain.Integration{}
	var configJSON []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, agent_id, type, name, config, template_id, created_at, updated_at
		 FROM integrations WHERE id = $1`, id,
	).Scan(&integration.ID, &integration.AgentID, &integration.Type, &integration.Name,
		&configJSON, &integration.TemplateID, &integration.CreatedAt, &integration.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrIntegrationNotFound
		}
		return nil, fmt.Errorf("querying integration: %w", err)
	}

	if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return integration, nil
}

func (r *IntegrationRepository) List(ctx context.Context, agentID string) ([]*domain.Integration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, agent_id, type, name, config, template_id, created_at, updated_at
		 FROM integrations WHERE agent_id = $1 ORDER BY created_at DESC`, agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*domain.Integration
	for rows.Next() {
		integration := &domain.Integration{}
		var configJSON []byte
		if err := rows.Scan(&integration.ID, &integration.AgentID, &integration.Type, &integration.Name,
			&configJSON, &integration.TemplateID, &integration.CreatedAt, &integration.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning integration: %w", err)
		}
		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			return nil, fmt.Errorf("unmarshaling config: %w", err)
		}
		integrations = append(integrations, integration)
	}
	return integrations, nil
}

func (r *IntegrationRepository) Update(ctx context.Context, integration *domain.Integration) error {
	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE integrations SET name = $2, config = $3
		 WHERE id = $1`,
		integration.ID, integration.Name, configJSON,
	)
	if err != nil {
		return fmt.Errorf("updating integration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrIntegrationNotFound
	}
	return nil
}

func (r *IntegrationRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM integrations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting integration: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrIntegrationNotFound
	}
	return nil
}
