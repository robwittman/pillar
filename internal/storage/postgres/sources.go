package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type SourceRepository struct {
	pool *pgxpool.Pool
}

func NewSourceRepository(pool *pgxpool.Pool) *SourceRepository {
	return &SourceRepository{pool: pool}
}

func (r *SourceRepository) Create(ctx context.Context, source *domain.Source) error {
	orgID := orgIDFromContext(ctx)
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sources (id, name, secret, org_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at, updated_at`,
		source.ID, source.Name, source.Secret, nullIfEmpty(orgID),
	).Scan(&source.CreatedAt, &source.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting source: %w", err)
	}
	return nil
}

func (r *SourceRepository) Get(ctx context.Context, id string) (*domain.Source, error) {
	s := &domain.Source{}
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, name, secret, created_at, updated_at FROM sources WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(&s.ID, &s.Name, &s.Secret, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSourceNotFound
		}
		return nil, fmt.Errorf("querying source: %w", err)
	}
	return s, nil
}

func (r *SourceRepository) List(ctx context.Context) ([]*domain.Source, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, name, secret, created_at, updated_at FROM sources`
	var args []any
	if orgID != "" {
		query += ` WHERE org_id = $1`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying sources: %w", err)
	}
	defer rows.Close()

	var sources []*domain.Source
	for rows.Next() {
		s := &domain.Source{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Secret, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (r *SourceRepository) Update(ctx context.Context, source *domain.Source) error {
	orgID := orgIDFromContext(ctx)
	query := `UPDATE sources SET name = $2, secret = $3 WHERE id = $1`
	args := []any{source.ID, source.Name, source.Secret}
	if orgID != "" {
		query += ` AND org_id = $4`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating source: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSourceNotFound
	}
	return nil
}

func (r *SourceRepository) Delete(ctx context.Context, id string) error {
	orgID := orgIDFromContext(ctx)
	query := `DELETE FROM sources WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting source: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSourceNotFound
	}
	return nil
}
