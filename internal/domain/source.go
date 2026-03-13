package domain

import (
	"context"
	"time"
)

type Source struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Secret    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SourceRepository interface {
	Create(ctx context.Context, source *Source) error
	Get(ctx context.Context, id string) (*Source, error)
	List(ctx context.Context) ([]*Source, error)
	Update(ctx context.Context, source *Source) error
	Delete(ctx context.Context, id string) error
}
