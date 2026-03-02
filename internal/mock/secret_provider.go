package mock

import "context"

type SecretProvider struct {
	PutFn    func(ctx context.Context, name string, value string) error
	GetFn    func(ctx context.Context, name string) (string, error)
	DeleteFn func(ctx context.Context, name string) error
}

func (m *SecretProvider) Put(ctx context.Context, name string, value string) error {
	return m.PutFn(ctx, name, value)
}

func (m *SecretProvider) Get(ctx context.Context, name string) (string, error) {
	return m.GetFn(ctx, name)
}

func (m *SecretProvider) Delete(ctx context.Context, name string) error {
	return m.DeleteFn(ctx, name)
}
