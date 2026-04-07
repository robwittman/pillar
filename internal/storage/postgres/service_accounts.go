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

type ServiceAccountRepository struct {
	pool *pgxpool.Pool
}

func NewServiceAccountRepository(pool *pgxpool.Pool) *ServiceAccountRepository {
	return &ServiceAccountRepository{pool: pool}
}

func (r *ServiceAccountRepository) Create(ctx context.Context, sa *domain.ServiceAccount) error {
	roles, err := json.Marshal(sa.Roles)
	if err != nil {
		return fmt.Errorf("marshaling roles: %w", err)
	}

	orgID := orgIDFromContext(ctx)
	err = r.pool.QueryRow(ctx,
		`INSERT INTO service_accounts (id, name, description, secret_hash, roles, disabled, org_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at, updated_at`,
		sa.ID, sa.Name, sa.Description, sa.SecretHash, roles, sa.Disabled, nullIfEmpty(orgID),
	).Scan(&sa.CreatedAt, &sa.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting service account: %w", err)
	}
	return nil
}

func (r *ServiceAccountRepository) Get(ctx context.Context, id string) (*domain.ServiceAccount, error) {
	return r.scanOne(ctx,
		`SELECT id, name, description, secret_hash, org_id, roles, disabled, created_at, updated_at
		 FROM service_accounts WHERE id = $1`, id)
}

func (r *ServiceAccountRepository) GetByName(ctx context.Context, name string) (*domain.ServiceAccount, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, name, description, secret_hash, org_id, roles, disabled, created_at, updated_at
		 FROM service_accounts WHERE name = $1`
	args := []any{name}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}
	return r.scanOneQuery(ctx, query, args...)
}

func (r *ServiceAccountRepository) List(ctx context.Context) ([]*domain.ServiceAccount, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, name, description, secret_hash, org_id, roles, disabled, created_at, updated_at FROM service_accounts`
	var args []any
	if orgID != "" {
		query += ` WHERE org_id = $1`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying service accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.ServiceAccount
	for rows.Next() {
		sa, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, sa)
	}
	return accounts, rows.Err()
}

func (r *ServiceAccountRepository) Update(ctx context.Context, sa *domain.ServiceAccount) error {
	roles, err := json.Marshal(sa.Roles)
	if err != nil {
		return fmt.Errorf("marshaling roles: %w", err)
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE service_accounts SET name = $2, description = $3, secret_hash = $4, roles = $5, disabled = $6
		 WHERE id = $1`,
		sa.ID, sa.Name, sa.Description, sa.SecretHash, roles, sa.Disabled,
	)
	if err != nil {
		return fmt.Errorf("updating service account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrServiceAccountNotFound
	}
	return nil
}

func (r *ServiceAccountRepository) Delete(ctx context.Context, id string) error {
	orgID := orgIDFromContext(ctx)
	query := `DELETE FROM service_accounts WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting service account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrServiceAccountNotFound
	}
	return nil
}

// scanOne queries by a single arg (typically ID) without org scoping.
// Used for Get-by-ID which needs to work across orgs for token resolution.
func (r *ServiceAccountRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.ServiceAccount, error) {
	return r.scanOneQuery(ctx, query, args...)
}

func (r *ServiceAccountRepository) scanOneQuery(ctx context.Context, query string, args ...any) (*domain.ServiceAccount, error) {
	sa := &domain.ServiceAccount{}
	var roles []byte

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&sa.ID, &sa.Name, &sa.Description, &sa.SecretHash, &sa.OrgID,
		&roles, &sa.Disabled, &sa.CreatedAt, &sa.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceAccountNotFound
		}
		return nil, fmt.Errorf("querying service account: %w", err)
	}
	if err := json.Unmarshal(roles, &sa.Roles); err != nil {
		return nil, fmt.Errorf("unmarshaling roles: %w", err)
	}
	return sa, nil
}

func (r *ServiceAccountRepository) scanRow(rows pgx.Rows) (*domain.ServiceAccount, error) {
	sa := &domain.ServiceAccount{}
	var roles []byte

	if err := rows.Scan(
		&sa.ID, &sa.Name, &sa.Description, &sa.SecretHash, &sa.OrgID,
		&roles, &sa.Disabled, &sa.CreatedAt, &sa.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scanning service account: %w", err)
	}
	if err := json.Unmarshal(roles, &sa.Roles); err != nil {
		return nil, fmt.Errorf("unmarshaling roles: %w", err)
	}
	return sa, nil
}
