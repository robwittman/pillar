package service

import (
	"context"
	"log/slog"
	"sync"
)

// AgentLogStore persists recent log entries per agent.
type AgentLogStore interface {
	Append(ctx context.Context, agentID string, entry string) error
	Query(ctx context.Context, agentID string, sinceNano int64, limit int) ([]string, error)
}

// LogSubscriber receives log entries in real-time.
type LogSubscriber struct {
	Ch      chan string
	AgentID string
}

// LogService handles agent log ingestion, storage, and real-time broadcasting.
type LogService struct {
	store  AgentLogStore
	logger *slog.Logger

	mu          sync.RWMutex
	subscribers map[string][]*LogSubscriber // agentID -> subscribers
}

func NewLogService(store AgentLogStore, logger *slog.Logger) *LogService {
	return &LogService{
		store:       store,
		logger:      logger,
		subscribers: make(map[string][]*LogSubscriber),
	}
}

// Ingest stores a log entry and broadcasts it to all live subscribers.
func (s *LogService) Ingest(ctx context.Context, agentID string, entry string) {
	if err := s.store.Append(ctx, agentID, entry); err != nil {
		s.logger.Warn("failed to store log entry", "agent_id", agentID, "error", err)
	}

	s.mu.RLock()
	subs := s.subscribers[agentID]
	s.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.Ch <- entry:
		default:
			// Drop if subscriber is slow
		}
	}
}

// Query returns historical log entries from the store.
func (s *LogService) Query(ctx context.Context, agentID string, sinceNano int64, limit int) ([]string, error) {
	return s.store.Query(ctx, agentID, sinceNano, limit)
}

// Subscribe creates a new real-time log subscriber for the given agent.
// The returned channel receives log entries as they arrive.
// Call Unsubscribe when done.
func (s *LogService) Subscribe(agentID string) *LogSubscriber {
	sub := &LogSubscriber{
		Ch:      make(chan string, 64),
		AgentID: agentID,
	}

	s.mu.Lock()
	s.subscribers[agentID] = append(s.subscribers[agentID], sub)
	s.mu.Unlock()

	return sub
}

// Unsubscribe removes a subscriber and closes its channel.
func (s *LogService) Unsubscribe(sub *LogSubscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs := s.subscribers[sub.AgentID]
	for i, existing := range subs {
		if existing == sub {
			s.subscribers[sub.AgentID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	close(sub.Ch)
}
