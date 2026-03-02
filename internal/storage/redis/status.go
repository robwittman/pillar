package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	heartbeatKeyPrefix = "pillar:agent:%s:heartbeat"
	onlineSetKey       = "pillar:agents:online"
)

type AgentStatusStore struct {
	client *redis.Client
}

func NewAgentStatusStore(client *redis.Client) *AgentStatusStore {
	return &AgentStatusStore{client: client}
}

func (s *AgentStatusStore) SetHeartbeat(ctx context.Context, agentID string, ttl time.Duration) error {
	key := fmt.Sprintf(heartbeatKeyPrefix, agentID)
	return s.client.Set(ctx, key, time.Now().Unix(), ttl).Err()
}

func (s *AgentStatusStore) IsOnline(ctx context.Context, agentID string) (bool, error) {
	score, err := s.client.ZScore(ctx, onlineSetKey, agentID).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return score > 0, nil
}

func (s *AgentStatusStore) SetOnline(ctx context.Context, agentID string) error {
	return s.client.ZAdd(ctx, onlineSetKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: agentID,
	}).Err()
}

func (s *AgentStatusStore) SetOffline(ctx context.Context, agentID string) error {
	key := fmt.Sprintf(heartbeatKeyPrefix, agentID)
	pipe := s.client.Pipeline()
	pipe.ZRem(ctx, onlineSetKey, agentID)
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *AgentStatusStore) ListOnline(ctx context.Context) ([]string, error) {
	return s.client.ZRevRange(ctx, onlineSetKey, 0, -1).Result()
}
