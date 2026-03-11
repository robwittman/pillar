package resolver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// Resolution is the result of resolving a plugin source.
type Resolution struct {
	BinaryPath string // absolute path to the cached binary
	Version    string // resolved version
}

// Resolver resolves a plugin source+version to a local binary path,
// downloading if necessary.
type Resolver interface {
	Resolve(ctx context.Context, source, version string) (*Resolution, error)
}

// CompositeResolver routes resolution requests to the appropriate source handler.
type CompositeResolver struct {
	github *GitHubSource
	logger *slog.Logger
}

// NewCompositeResolver creates a resolver that handles GitHub and HTTP sources.
func NewCompositeResolver(cache *Cache, logger *slog.Logger) *CompositeResolver {
	return &CompositeResolver{
		github: NewGitHubSource(cache, logger),
		logger: logger,
	}
}

func (r *CompositeResolver) Resolve(ctx context.Context, source, version string) (*Resolution, error) {
	if strings.HasPrefix(source, "github.com/") {
		return r.github.Resolve(ctx, source, version)
	}
	return nil, fmt.Errorf("unsupported plugin source: %s (supported: github.com/owner/repo)", source)
}
