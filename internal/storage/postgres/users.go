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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	roles, err := json.Marshal(user.Roles)
	if err != nil {
		return fmt.Errorf("marshaling roles: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (id, email, display_name, password_hash, provider, provider_sub_id, roles, disabled)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING created_at, updated_at`,
		user.ID, user.Email, user.DisplayName, user.PasswordHash,
		user.Provider, user.ProviderSubID, roles, user.Disabled,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *UserRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	return r.scanOne(ctx,
		`SELECT id, email, display_name, password_hash, provider, provider_sub_id, roles, disabled, created_at, updated_at
		 FROM users WHERE id = $1`, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scanOne(ctx,
		`SELECT id, email, display_name, password_hash, provider, provider_sub_id, roles, disabled, created_at, updated_at
		 FROM users WHERE email = $1`, email)
}

func (r *UserRepository) GetByProviderSub(ctx context.Context, provider, subID string) (*domain.User, error) {
	return r.scanOne(ctx,
		`SELECT id, email, display_name, password_hash, provider, provider_sub_id, roles, disabled, created_at, updated_at
		 FROM users WHERE provider = $1 AND provider_sub_id = $2`, provider, subID)
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, display_name, password_hash, provider, provider_sub_id, roles, disabled, created_at, updated_at
		 FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	roles, err := json.Marshal(user.Roles)
	if err != nil {
		return fmt.Errorf("marshaling roles: %w", err)
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET email = $2, display_name = $3, password_hash = $4,
		 provider = $5, provider_sub_id = $6, roles = $7, disabled = $8
		 WHERE id = $1`,
		user.ID, user.Email, user.DisplayName, user.PasswordHash,
		user.Provider, user.ProviderSubID, roles, user.Disabled,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.User, error) {
	user := &domain.User{}
	var roles []byte

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash,
		&user.Provider, &user.ProviderSubID, &roles, &user.Disabled,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user: %w", err)
	}
	if err := json.Unmarshal(roles, &user.Roles); err != nil {
		return nil, fmt.Errorf("unmarshaling roles: %w", err)
	}
	return user, nil
}

func (r *UserRepository) scanRow(rows pgx.Rows) (*domain.User, error) {
	user := &domain.User{}
	var roles []byte

	if err := rows.Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.PasswordHash,
		&user.Provider, &user.ProviderSubID, &roles, &user.Disabled,
		&user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scanning user: %w", err)
	}
	if err := json.Unmarshal(roles, &user.Roles); err != nil {
		return nil, fmt.Errorf("unmarshaling roles: %w", err)
	}
	return user, nil
}
