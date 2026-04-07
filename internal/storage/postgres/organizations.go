package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type OrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{pool: pool}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO organizations (id, name, slug, personal, owner_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at, updated_at`,
		org.ID, org.Name, org.Slug, org.Personal, org.OwnerID,
	).Scan(&org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting organization: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) Get(ctx context.Context, id string) (*domain.Organization, error) {
	return r.scanOne(ctx,
		`SELECT id, name, slug, personal, owner_id, created_at, updated_at
		 FROM organizations WHERE id = $1`, id)
}

func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	return r.scanOne(ctx,
		`SELECT id, name, slug, personal, owner_id, created_at, updated_at
		 FROM organizations WHERE slug = $1`, slug)
}

func (r *OrganizationRepository) GetPersonalOrg(ctx context.Context, ownerID string) (*domain.Organization, error) {
	return r.scanOne(ctx,
		`SELECT id, name, slug, personal, owner_id, created_at, updated_at
		 FROM organizations WHERE owner_id = $1 AND personal = true`, ownerID)
}

func (r *OrganizationRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Organization, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT o.id, o.name, o.slug, o.personal, o.owner_id, o.created_at, o.updated_at
		 FROM organizations o
		 JOIN memberships m ON m.org_id = o.id
		 WHERE m.user_id = $1
		 ORDER BY o.personal DESC, o.name ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("querying organizations: %w", err)
	}
	defer rows.Close()

	var orgs []*domain.Organization
	for rows.Next() {
		org, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

func (r *OrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE organizations SET name = $2, slug = $3 WHERE id = $1`,
		org.ID, org.Name, org.Slug,
	)
	if err != nil {
		return fmt.Errorf("updating organization: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrgNotFound
	}
	return nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting organization: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrgNotFound
	}
	return nil
}

func (r *OrganizationRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.Organization, error) {
	org := &domain.Organization{}
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Personal, &org.OwnerID,
		&org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrgNotFound
		}
		return nil, fmt.Errorf("querying organization: %w", err)
	}
	return org, nil
}

func (r *OrganizationRepository) scanRow(rows pgx.Rows) (*domain.Organization, error) {
	org := &domain.Organization{}
	if err := rows.Scan(
		&org.ID, &org.Name, &org.Slug, &org.Personal, &org.OwnerID,
		&org.CreatedAt, &org.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scanning organization: %w", err)
	}
	return org, nil
}
