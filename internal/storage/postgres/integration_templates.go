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

type IntegrationTemplateRepository struct {
	pool *pgxpool.Pool
}

func NewIntegrationTemplateRepository(pool *pgxpool.Pool) *IntegrationTemplateRepository {
	return &IntegrationTemplateRepository{pool: pool}
}

func (r *IntegrationTemplateRepository) Create(ctx context.Context, template *domain.IntegrationTemplate) error {
	configJSON, err := json.Marshal(template.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	selectorJSON, err := json.Marshal(template.Selector)
	if err != nil {
		return fmt.Errorf("marshaling selector: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO integration_templates (id, type, name, config, selector)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at, updated_at`,
		template.ID, template.Type, template.Name, configJSON, selectorJSON,
	).Scan(&template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrInvalidIntegrationTemplate
		}
		return fmt.Errorf("inserting integration template: %w", err)
	}
	return nil
}

func (r *IntegrationTemplateRepository) Get(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
	template := &domain.IntegrationTemplate{}
	var configJSON, selectorJSON []byte

	err := r.pool.QueryRow(ctx,
		`SELECT id, type, name, config, selector, created_at, updated_at
		 FROM integration_templates WHERE id = $1`, id,
	).Scan(&template.ID, &template.Type, &template.Name,
		&configJSON, &selectorJSON, &template.CreatedAt, &template.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrIntegrationTemplateNotFound
		}
		return nil, fmt.Errorf("querying integration template: %w", err)
	}

	if err := json.Unmarshal(configJSON, &template.Config); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	if err := json.Unmarshal(selectorJSON, &template.Selector); err != nil {
		return nil, fmt.Errorf("unmarshaling selector: %w", err)
	}
	return template, nil
}

func (r *IntegrationTemplateRepository) List(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, type, name, config, selector, created_at, updated_at
		 FROM integration_templates ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying integration templates: %w", err)
	}
	defer rows.Close()

	var templates []*domain.IntegrationTemplate
	for rows.Next() {
		template := &domain.IntegrationTemplate{}
		var configJSON, selectorJSON []byte
		if err := rows.Scan(&template.ID, &template.Type, &template.Name,
			&configJSON, &selectorJSON, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning integration template: %w", err)
		}
		if err := json.Unmarshal(configJSON, &template.Config); err != nil {
			return nil, fmt.Errorf("unmarshaling config: %w", err)
		}
		if err := json.Unmarshal(selectorJSON, &template.Selector); err != nil {
			return nil, fmt.Errorf("unmarshaling selector: %w", err)
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func (r *IntegrationTemplateRepository) Update(ctx context.Context, template *domain.IntegrationTemplate) error {
	configJSON, err := json.Marshal(template.Config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	selectorJSON, err := json.Marshal(template.Selector)
	if err != nil {
		return fmt.Errorf("marshaling selector: %w", err)
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE integration_templates SET name = $2, config = $3, selector = $4
		 WHERE id = $1`,
		template.ID, template.Name, configJSON, selectorJSON,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrInvalidIntegrationTemplate
		}
		return fmt.Errorf("updating integration template: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrIntegrationTemplateNotFound
	}
	return nil
}

func (r *IntegrationTemplateRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM integration_templates WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting integration template: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrIntegrationTemplateNotFound
	}
	return nil
}

func (r *IntegrationTemplateRepository) FindMatchingTemplates(ctx context.Context, labels map[string]string) ([]*domain.IntegrationTemplate, error) {
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("marshaling labels: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, type, name, config, selector, created_at, updated_at
		 FROM integration_templates WHERE $1::jsonb @> selector
		 ORDER BY created_at DESC`, labelsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("querying matching templates: %w", err)
	}
	defer rows.Close()

	var templates []*domain.IntegrationTemplate
	for rows.Next() {
		template := &domain.IntegrationTemplate{}
		var configJSON, selectorJSON []byte
		if err := rows.Scan(&template.ID, &template.Type, &template.Name,
			&configJSON, &selectorJSON, &template.CreatedAt, &template.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning integration template: %w", err)
		}
		if err := json.Unmarshal(configJSON, &template.Config); err != nil {
			return nil, fmt.Errorf("unmarshaling config: %w", err)
		}
		if err := json.Unmarshal(selectorJSON, &template.Selector); err != nil {
			return nil, fmt.Errorf("unmarshaling selector: %w", err)
		}
		templates = append(templates, template)
	}
	return templates, nil
}
