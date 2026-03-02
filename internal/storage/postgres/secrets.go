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
	_, err := s.pool.Exec(ctx,
		`INSERT INTO agent_secrets (name, value) VALUES ($1, $2)
		 ON CONFLICT (name) DO UPDATE SET value = $2`,
		name, value,
	)
	if err != nil {
		return fmt.Errorf("storing secret: %w", err)
	}
	return nil
}

func (s *SecretStore) Get(ctx context.Context, name string) (string, error) {
	var value string
	err := s.pool.QueryRow(ctx,
		`SELECT value FROM agent_secrets WHERE name = $1`, name,
	).Scan(&value)
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
