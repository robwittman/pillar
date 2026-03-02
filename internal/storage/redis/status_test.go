//go:build integration

package redis

import (
	"context"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedis(t *testing.T) (*goredis.Client, func()) {
	t.Helper()
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp").WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	endpoint, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	client := goredis.NewClient(&goredis.Options{Addr: endpoint})
	require.NoError(t, client.Ping(ctx).Err())

	return client, func() {
		client.Close()
		redisContainer.Terminate(ctx)
	}
}

func TestAgentStatusStore_SetHeartbeatAndTTL(t *testing.T) {
	client, cleanup := setupRedis(t)
	defer cleanup()

	store := NewAgentStatusStore(client)
	ctx := context.Background()

	err := store.SetHeartbeat(ctx, "agent1", 2*time.Second)
	require.NoError(t, err)

	// Key should exist
	ttl := client.TTL(ctx, "pillar:agent:agent1:heartbeat").Val()
	assert.True(t, ttl > 0 && ttl <= 2*time.Second)

	// Wait for expiry
	time.Sleep(3 * time.Second)
	exists := client.Exists(ctx, "pillar:agent:agent1:heartbeat").Val()
	assert.Equal(t, int64(0), exists)
}

func TestAgentStatusStore_SetOnlineAndIsOnline(t *testing.T) {
	client, cleanup := setupRedis(t)
	defer cleanup()

	store := NewAgentStatusStore(client)
	ctx := context.Background()

	online, err := store.IsOnline(ctx, "agent1")
	require.NoError(t, err)
	assert.False(t, online)

	require.NoError(t, store.SetOnline(ctx, "agent1"))

	online, err = store.IsOnline(ctx, "agent1")
	require.NoError(t, err)
	assert.True(t, online)
}

func TestAgentStatusStore_SetOffline(t *testing.T) {
	client, cleanup := setupRedis(t)
	defer cleanup()

	store := NewAgentStatusStore(client)
	ctx := context.Background()

	require.NoError(t, store.SetOnline(ctx, "agent1"))
	require.NoError(t, store.SetHeartbeat(ctx, "agent1", 30*time.Second))

	require.NoError(t, store.SetOffline(ctx, "agent1"))

	online, err := store.IsOnline(ctx, "agent1")
	require.NoError(t, err)
	assert.False(t, online)

	exists := client.Exists(ctx, "pillar:agent:agent1:heartbeat").Val()
	assert.Equal(t, int64(0), exists)
}

func TestAgentStatusStore_ListOnline(t *testing.T) {
	client, cleanup := setupRedis(t)
	defer cleanup()

	store := NewAgentStatusStore(client)
	ctx := context.Background()

	require.NoError(t, store.SetOnline(ctx, "agent-a"))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, store.SetOnline(ctx, "agent-b"))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, store.SetOnline(ctx, "agent-c"))

	ids, err := store.ListOnline(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 3)
	// ZRevRange => most recent first
	assert.Equal(t, "agent-c", ids[0])
	assert.Equal(t, "agent-b", ids[1])
	assert.Equal(t, "agent-a", ids[2])
}
