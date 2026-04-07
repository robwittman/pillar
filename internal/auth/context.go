package auth

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type principalKey struct{}
type orgKey struct{}

// ContextWithPrincipal returns a new context with the given principal attached.
func ContextWithPrincipal(ctx context.Context, p *domain.Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFromContext extracts the authenticated principal from context.
func PrincipalFromContext(ctx context.Context) (*domain.Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(*domain.Principal)
	return p, ok
}

// ContextWithOrg returns a new context with the given org context attached.
func ContextWithOrg(ctx context.Context, oc *domain.OrgContext) context.Context {
	return context.WithValue(ctx, orgKey{}, oc)
}

// OrgFromContext extracts the organization context from context.
func OrgFromContext(ctx context.Context) (*domain.OrgContext, bool) {
	oc, ok := ctx.Value(orgKey{}).(*domain.OrgContext)
	return oc, ok
}
