//go:build integration

package postgres

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/robwittman/pillar/internal/domain"
)

func setupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "migrations")

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pillar_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(
			filepath.Join(migrationsDir, "001_create_agents.up.sql"),
			filepath.Join(migrationsDir, "002_create_agent_configs.up.sql"),
		),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	return pool, func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}
}

func TestAgentRepository_CreateAndGet(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewAgentRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "test-id-1",
		Name:     "test-agent",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{"env": "test"},
		Labels:   map[string]string{"tier": "1"},
	}

	err := repo.Create(ctx, agent)
	require.NoError(t, err)
	assert.False(t, agent.CreatedAt.IsZero())
	assert.False(t, agent.UpdatedAt.IsZero())

	got, err := repo.Get(ctx, "test-id-1")
	require.NoError(t, err)
	assert.Equal(t, "test-agent", got.Name)
	assert.Equal(t, domain.AgentStatusPending, got.Status)
	assert.Equal(t, "test", got.Metadata["env"])
	assert.Equal(t, "1", got.Labels["tier"])
}

func TestAgentRepository_List(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewAgentRepository(pool)
	ctx := context.Background()

	for i, name := range []string{"agent-a", "agent-b", "agent-c"} {
		err := repo.Create(ctx, &domain.Agent{
			ID:       name,
			Name:     name,
			Status:   domain.AgentStatusPending,
			Metadata: map[string]string{},
			Labels:   map[string]string{},
		})
		require.NoError(t, err)
		if i < 2 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	agents, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, agents, 3)
	// ORDER BY created_at DESC
	assert.Equal(t, "agent-c", agents[0].ID)
}

func TestAgentRepository_Update(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewAgentRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "upd-1",
		Name:     "original",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, repo.Create(ctx, agent))

	agent.Name = "updated"
	agent.Metadata = map[string]string{"new": "val"}
	require.NoError(t, repo.Update(ctx, agent))

	got, err := repo.Get(ctx, "upd-1")
	require.NoError(t, err)
	assert.Equal(t, "updated", got.Name)
	assert.Equal(t, "val", got.Metadata["new"])
}

func TestAgentRepository_Delete(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewAgentRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "del-1",
		Name:     "to-delete",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, repo.Create(ctx, agent))
	require.NoError(t, repo.Delete(ctx, "del-1"))

	_, err := repo.Get(ctx, "del-1")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

func TestAgentRepository_UpdateStatus(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewAgentRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "status-1",
		Name:     "agent",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, repo.Create(ctx, agent))
	require.NoError(t, repo.UpdateStatus(ctx, "status-1", domain.AgentStatusRunning))

	got, err := repo.Get(ctx, "status-1")
	require.NoError(t, err)
	assert.Equal(t, domain.AgentStatusRunning, got.Status)
}

func TestAgentRepository_NotFound(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewAgentRepository(pool)
	ctx := context.Background()

	_, err := repo.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)

	err = repo.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)

	err = repo.UpdateStatus(ctx, "nonexistent", domain.AgentStatusRunning)
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}
