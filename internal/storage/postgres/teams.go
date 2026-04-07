package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type TeamRepository struct {
	pool *pgxpool.Pool
}

func NewTeamRepository(pool *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{pool: pool}
}

func (r *TeamRepository) Create(ctx context.Context, t *domain.Team) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO teams (id, org_id, name)
		 VALUES ($1, $2, $3)
		 RETURNING created_at, updated_at`,
		t.ID, t.OrgID, t.Name,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting team: %w", err)
	}
	return nil
}

func (r *TeamRepository) Get(ctx context.Context, id string) (*domain.Team, error) {
	t := &domain.Team{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, org_id, name, created_at, updated_at
		 FROM teams WHERE id = $1`, id,
	).Scan(&t.ID, &t.OrgID, &t.Name, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("querying team: %w", err)
	}
	return t, nil
}

func (r *TeamRepository) ListByOrg(ctx context.Context, orgID string) ([]*domain.Team, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, name, created_at, updated_at
		 FROM teams WHERE org_id = $1 ORDER BY name ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("querying teams: %w", err)
	}
	defer rows.Close()

	var teams []*domain.Team
	for rows.Next() {
		t := &domain.Team{}
		if err := rows.Scan(&t.ID, &t.OrgID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *TeamRepository) Update(ctx context.Context, t *domain.Team) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE teams SET name = $2 WHERE id = $1`,
		t.ID, t.Name,
	)
	if err != nil {
		return fmt.Errorf("updating team: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTeamNotFound
	}
	return nil
}

func (r *TeamRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM teams WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting team: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTeamNotFound
	}
	return nil
}
