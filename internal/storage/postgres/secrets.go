package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type SecretStore struct {
	pool *pgxpool.Pool
}

func NewSecretStore(pool *pgxpool.Pool) *SecretStore {
	return &SecretStore{pool: pool}
}

func (s *SecretStore) Put(ctx context.Context, name string, value string) error {
	orgID := orgIDFromContext(ctx)
	_, err := s.pool.Exec(ctx,
		`INSERT INTO agent_secrets (name, value, org_id) VALUES ($1, $2, $3)
		 ON CONFLICT (name) DO UPDATE SET value = $2`,
		name, value, nullIfEmpty(orgID),
	)
	if err != nil {
		return fmt.Errorf("storing secret: %w", err)
	}
	return nil
}

func (s *SecretStore) Get(ctx context.Context, name string) (string, error) {
	var value string
	orgID := orgIDFromContext(ctx)
	query := `SELECT value FROM agent_secrets WHERE name = $1`
	args := []any{name}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	err := s.pool.QueryRow(ctx, query, args...).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrSecretNotFound
		}
		return "", fmt.Errorf("querying secret: %w", err)
	}
	return value, nil
}

func (s *SecretStore) Delete(ctx context.Context, name string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM agent_secrets WHERE name = $1`, name)
	if err != nil {
		return fmt.Errorf("deleting secret: %w", err)
	}
	return nil
}
