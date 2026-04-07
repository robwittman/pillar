package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type MembershipRepository struct {
	pool *pgxpool.Pool
}

func NewMembershipRepository(pool *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{pool: pool}
}

func (r *MembershipRepository) Create(ctx context.Context, m *domain.Membership) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO memberships (id, org_id, user_id, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at, updated_at`,
		m.ID, m.OrgID, m.UserID, m.Role,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting membership: %w", err)
	}
	return nil
}

func (r *MembershipRepository) Get(ctx context.Context, id string) (*domain.Membership, error) {
	return r.scanOne(ctx,
		`SELECT id, org_id, user_id, role, created_at, updated_at
		 FROM memberships WHERE id = $1`, id)
}

func (r *MembershipRepository) GetByOrgAndUser(ctx context.Context, orgID, userID string) (*domain.Membership, error) {
	return r.scanOne(ctx,
		`SELECT id, org_id, user_id, role, created_at, updated_at
		 FROM memberships WHERE org_id = $1 AND user_id = $2`, orgID, userID)
}

func (r *MembershipRepository) ListByOrg(ctx context.Context, orgID string) ([]*domain.Membership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, user_id, role, created_at, updated_at
		 FROM memberships WHERE org_id = $1 ORDER BY created_at ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("querying memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.Membership
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}

func (r *MembershipRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Membership, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, user_id, role, created_at, updated_at
		 FROM memberships WHERE user_id = $1 ORDER BY created_at ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("querying memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.Membership
	for rows.Next() {
		m, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}

func (r *MembershipRepository) Update(ctx context.Context, m *domain.Membership) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE memberships SET role = $2 WHERE id = $1`,
		m.ID, m.Role,
	)
	if err != nil {
		return fmt.Errorf("updating membership: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMembershipNotFound
	}
	return nil
}

func (r *MembershipRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM memberships WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting membership: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMembershipNotFound
	}
	return nil
}

func (r *MembershipRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Membership, error) {
	m := &domain.Membership{}
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMembershipNotFound
		}
		return nil, fmt.Errorf("querying membership: %w", err)
	}
	return m, nil
}

func (r *MembershipRepository) scanRow(rows pgx.Rows) (*domain.Membership, error) {
	m := &domain.Membership{}
	if err := rows.Scan(
		&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scanning membership: %w", err)
	}
	return m, nil
}
