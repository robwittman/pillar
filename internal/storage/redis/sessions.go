package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robwittman/pillar/internal/domain"
)

const sessionKeyPrefix = "pillar:session:%s"

type SessionStore struct {
	client *redis.Client
}

func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{client: client}
}

func (s *SessionStore) Create(ctx context.Context, session *domain.Session) error {
	key := fmt.Sprintf(sessionKeyPrefix, session.ID)

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return domain.ErrSessionExpired
	}

	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("storing session: %w", err)
	}
	return nil
}

func (s *SessionStore) Get(ctx context.Context, id string) (*domain.Session, error) {
	key := fmt.Sprintf(sessionKeyPrefix, id)

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("retrieving session: %w", err)
	}

	session := &domain.Session{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil, fmt.Errorf("unmarshaling session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.client.Del(ctx, key).Err()
		return nil, domain.ErrSessionExpired
	}

	return session, nil
}

func (s *SessionStore) Delete(ctx context.Context, id string) error {
	key := fmt.Sprintf(sessionKeyPrefix, id)
	return s.client.Del(ctx, key).Err()
}
