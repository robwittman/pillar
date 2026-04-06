package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type APITokenRepository struct {
	pool *pgxpool.Pool
}

func NewAPITokenRepository(pool *pgxpool.Pool) *APITokenRepository {
	return &APITokenRepository{pool: pool}
}

func (r *APITokenRepository) Create(ctx context.Context, token *domain.APIToken) error {
	scopes, err := json.Marshal(token.Scopes)
	if err != nil {
		return fmt.Errorf("marshaling scopes: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`INSERT INTO api_tokens (id, name, token_hash, owner_id, owner_type, scopes, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at`,
		token.ID, token.Name, token.TokenHash, token.OwnerID, token.OwnerType,
		scopes, token.ExpiresAt,
	).Scan(&token.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting api token: %w", err)
	}
	return nil
}

func (r *APITokenRepository) Get(ctx context.Context, id string) (*domain.APIToken, error) {
	return r.scanOne(ctx,
		`SELECT id, name, token_hash, owner_id, owner_type, scopes, expires_at, last_used_at, created_at
		 FROM api_tokens WHERE id = $1`, id)
}

func (r *APITokenRepository) GetByHash(ctx context.Context, hash string) (*domain.APIToken, error) {
	return r.scanOne(ctx,
		`SELECT id, name, token_hash, owner_id, owner_type, scopes, expires_at, last_used_at, created_at
		 FROM api_tokens WHERE token_hash = $1`, hash)
}

func (r *APITokenRepository) ListByOwner(ctx context.Context, ownerID string, ownerType domain.PrincipalType) ([]*domain.APIToken, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, token_hash, owner_id, owner_type, scopes, expires_at, last_used_at, created_at
		 FROM api_tokens WHERE owner_id = $1 AND owner_type = $2
		 ORDER BY created_at DESC`, ownerID, ownerType)
	if err != nil {
		return nil, fmt.Errorf("querying api tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*domain.APIToken
	for rows.Next() {
		token, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *APITokenRepository) UpdateLastUsed(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE api_tokens SET last_used_at = $2 WHERE id = $1`,
		id, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("updating token last used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}

func (r *APITokenRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM api_tokens WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting api token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}

func (r *APITokenRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.APIToken, error) {
	token := &domain.APIToken{}
	var scopes []byte

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&token.ID, &token.Name, &token.TokenHash, &token.OwnerID, &token.OwnerType,
		&scopes, &token.ExpiresAt, &token.LastUsedAt, &token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("querying api token: %w", err)
	}
	if err := json.Unmarshal(scopes, &token.Scopes); err != nil {
		return nil, fmt.Errorf("unmarshaling scopes: %w", err)
	}
	return token, nil
}

func (r *APITokenRepository) scanRow(rows pgx.Rows) (*domain.APIToken, error) {
	token := &domain.APIToken{}
	var scopes []byte

	if err := rows.Scan(
		&token.ID, &token.Name, &token.TokenHash, &token.OwnerID, &token.OwnerType,
		&scopes, &token.ExpiresAt, &token.LastUsedAt, &token.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("scanning api token: %w", err)
	}
	if err := json.Unmarshal(scopes, &token.Scopes); err != nil {
		return nil, fmt.Errorf("unmarshaling scopes: %w", err)
	}
	return token, nil
}
