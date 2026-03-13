package domain

import (
	"context"
	"time"
)

type FilterCondition struct {
	Path  string `json:"path"`
	Op    string `json:"op"`
	Value string `json:"value,omitempty"`
}

type TriggerFilter struct {
	Conditions []FilterCondition `json:"conditions"`
}

type Trigger struct {
	ID           string        `json:"id"`
	SourceID     string        `json:"source_id"`
	AgentID      string        `json:"agent_id"`
	Name         string        `json:"name"`
	Filter       TriggerFilter `json:"filter"`
	TaskTemplate string        `json:"task_template"`
	Enabled      bool          `json:"enabled"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type TriggerRepository interface {
	Create(ctx context.Context, trigger *Trigger) error
	Get(ctx context.Context, id string) (*Trigger, error)
	List(ctx context.Context) ([]*Trigger, error)
	ListBySource(ctx context.Context, sourceID string) ([]*Trigger, error)
	Update(ctx context.Context, trigger *Trigger) error
	Delete(ctx context.Context, id string) error
}
