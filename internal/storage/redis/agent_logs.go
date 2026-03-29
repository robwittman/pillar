package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	logKeyPrefix  = "pillar:agent:%s:logs"
	maxLogEntries = 1000
	logTTL        = 24 * time.Hour
)

// AgentLogStore stores recent agent log entries in Redis using a sorted set.
// Entries are scored by Unix timestamp (nanoseconds) for ordering.
type AgentLogStore struct {
	client *redis.Client
}

func NewAgentLogStore(client *redis.Client) *AgentLogStore {
	return &AgentLogStore{client: client}
}

// Append adds a log entry for the given agent. The entry is stored as a
// JSON string scored by the current timestamp in nanoseconds.
func (s *AgentLogStore) Append(ctx context.Context, agentID string, entry string) error {
	key := fmt.Sprintf(logKeyPrefix, agentID)
	score := float64(time.Now().UnixNano())

	pipe := s.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: entry})
	// Trim to keep only the most recent entries
	pipe.ZRemRangeByRank(ctx, key, 0, int64(-maxLogEntries-1))
	// Reset TTL on each write
	pipe.Expire(ctx, key, logTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// Query returns log entries for the given agent, ordered by time.
// sinceNano filters entries after the given Unix nanosecond timestamp (0 = all).
// limit caps the number of returned entries (0 = default 200).
func (s *AgentLogStore) Query(ctx context.Context, agentID string, sinceNano int64, limit int) ([]string, error) {
	key := fmt.Sprintf(logKeyPrefix, agentID)

	if limit <= 0 {
		limit = 200
	}

	minScore := "-inf"
	if sinceNano > 0 {
		minScore = strconv.FormatInt(sinceNano, 10)
	}

	return s.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:   minScore,
		Max:   "+inf",
		Count: int64(limit),
	}).Result()
}
