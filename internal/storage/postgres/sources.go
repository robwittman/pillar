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
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sources (id, name, secret)
		 VALUES ($1, $2, $3)
		 RETURNING created_at, updated_at`,
		source.ID, source.Name, source.Secret,
	).Scan(&source.CreatedAt, &source.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting source: %w", err)
	}
	return nil
}

func (r *SourceRepository) Get(ctx context.Context, id string) (*domain.Source, error) {
	s := &domain.Source{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, secret, created_at, updated_at FROM sources WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.Secret, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSourceNotFound
		}
		return nil, fmt.Errorf("querying source: %w", err)
	}
	return s, nil
}

func (r *SourceRepository) List(ctx context.Context) ([]*domain.Source, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, secret, created_at, updated_at FROM sources ORDER BY created_at DESC`)
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
	tag, err := r.pool.Exec(ctx,
		`UPDATE sources SET name = $2, secret = $3 WHERE id = $1`,
		source.ID, source.Name, source.Secret,
	)
	if err != nil {
		return fmt.Errorf("updating source: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSourceNotFound
	}
	return nil
}

func (r *SourceRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sources WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting source: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSourceNotFound
	}
	return nil
}
