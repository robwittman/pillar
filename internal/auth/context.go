package auth

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type contextKey struct{}

// ContextWithPrincipal returns a new context with the given principal attached.
func ContextWithPrincipal(ctx context.Context, p *domain.Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// PrincipalFromContext extracts the authenticated principal from context.
func PrincipalFromContext(ctx context.Context) (*domain.Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(*domain.Principal)
	return p, ok
}
