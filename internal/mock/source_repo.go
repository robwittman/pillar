package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type SourceRepository struct {
	CreateFn func(ctx context.Context, source *domain.Source) error
	GetFn    func(ctx context.Context, id string) (*domain.Source, error)
	ListFn   func(ctx context.Context) ([]*domain.Source, error)
	UpdateFn func(ctx context.Context, source *domain.Source) error
	DeleteFn func(ctx context.Context, id string) error
}

func (m *SourceRepository) Create(ctx context.Context, source *domain.Source) error {
	return m.CreateFn(ctx, source)
}

func (m *SourceRepository) Get(ctx context.Context, id string) (*domain.Source, error) {
	return m.GetFn(ctx, id)
}

func (m *SourceRepository) List(ctx context.Context) ([]*domain.Source, error) {
	return m.ListFn(ctx)
}

func (m *SourceRepository) Update(ctx context.Context, source *domain.Source) error {
	return m.UpdateFn(ctx, source)
}

func (m *SourceRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
