package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/auth"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return pool, nil
}

// orgIDFromContext extracts the org ID from the request context.
// Returns empty string when no org context is present (auth disabled).
func orgIDFromContext(ctx context.Context) string {
	oc, ok := auth.OrgFromContext(ctx)
	if !ok {
		return ""
	}
	return oc.OrgID
}
