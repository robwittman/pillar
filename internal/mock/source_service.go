package mock

import (
	"context"
	"encoding/json"

	"github.com/robwittman/pillar/internal/domain"
)

type SourceService struct {
	CreateFn       func(ctx context.Context, name string) (*domain.Source, string, error)
	GetFn          func(ctx context.Context, id string) (*domain.Source, error)
	ListFn         func(ctx context.Context) ([]*domain.Source, error)
	UpdateFn       func(ctx context.Context, id string, name string) (*domain.Source, error)
	DeleteFn       func(ctx context.Context, id string) error
	RotateSecretFn func(ctx context.Context, id string) (*domain.Source, string, error)
	HandleEventFn  func(ctx context.Context, sourceID string, signature string, payload json.RawMessage) ([]string, error)
}

func (m *SourceService) Create(ctx context.Context, name string) (*domain.Source, string, error) {
	return m.CreateFn(ctx, name)
}

func (m *SourceService) Get(ctx context.Context, id string) (*domain.Source, error) {
	return m.GetFn(ctx, id)
}

func (m *SourceService) List(ctx context.Context) ([]*domain.Source, error) {
	return m.ListFn(ctx)
}

func (m *SourceService) Update(ctx context.Context, id string, name string) (*domain.Source, error) {
	return m.UpdateFn(ctx, id, name)
}

func (m *SourceService) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *SourceService) RotateSecret(ctx context.Context, id string) (*domain.Source, string, error) {
	return m.RotateSecretFn(ctx, id)
}

func (m *SourceService) HandleEvent(ctx context.Context, sourceID string, signature string, payload json.RawMessage) ([]string, error) {
	return m.HandleEventFn(ctx, sourceID, signature, payload)
}
