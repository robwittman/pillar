package auth

import (
	"context"
	"fmt"

	"github.com/robwittman/pillar/internal/config"
	"github.com/robwittman/pillar/internal/domain"
)

// ProviderInfo is the public metadata for a configured identity provider,
// returned to the UI so it can render the login page.
type ProviderInfo struct {
	Name    string       `json:"name"`
	Type    ProviderType `json:"type"`
	AuthURL string       `json:"auth_url,omitempty"`
}

// ProviderRegistry holds all configured identity providers.
type ProviderRegistry struct {
	providers map[string]IdentityProvider
}

// NewProviderRegistry builds the registry from configuration.
// It requires a UserRepository only when a local provider is configured.
func NewProviderRegistry(ctx context.Context, configs []config.AuthProviderConfig, userRepo domain.UserRepository) (*ProviderRegistry, error) {
	r := &ProviderRegistry{
		providers: make(map[string]IdentityProvider),
	}

	for _, cfg := range configs {
		if cfg.Name == "" {
			return nil, fmt.Errorf("auth provider config missing name")
		}
		if _, exists := r.providers[cfg.Name]; exists {
			return nil, fmt.Errorf("duplicate auth provider name: %s", cfg.Name)
		}

		switch ProviderType(cfg.Type) {
		case ProviderLocal:
			if userRepo == nil {
				return nil, fmt.Errorf("local provider requires a UserRepository")
			}
			r.providers[cfg.Name] = NewLocalProvider(userRepo)

		case ProviderOIDC:
			p, err := NewOIDCProvider(ctx, OIDCProviderConfig{
				Name:         cfg.Name,
				IssuerURL:    cfg.IssuerURL,
				ClientID:     cfg.ClientID,
				ClientSecret: cfg.ClientSecret,
				RedirectURL:  cfg.RedirectURL,
				Scopes:       cfg.Scopes,
			})
			if err != nil {
				return nil, fmt.Errorf("initializing OIDC provider %s: %w", cfg.Name, err)
			}
			r.providers[cfg.Name] = p

		case ProviderGitHub:
			r.providers[cfg.Name] = NewGitHubProvider(GitHubProviderConfig{
				Name:         cfg.Name,
				ClientID:     cfg.ClientID,
				ClientSecret: cfg.ClientSecret,
				RedirectURL:  cfg.RedirectURL,
				Scopes:       cfg.Scopes,
			})

		default:
			return nil, fmt.Errorf("unknown auth provider type: %s", cfg.Type)
		}
	}

	return r, nil
}

// Get returns a provider by name.
func (r *ProviderRegistry) Get(name string) (IdentityProvider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// List returns info about all configured providers.
func (r *ProviderRegistry) List() []ProviderInfo {
	infos := make([]ProviderInfo, 0, len(r.providers))
	for _, p := range r.providers {
		info := ProviderInfo{
			Name: p.Name(),
			Type: p.Type(),
		}
		if url := p.AuthCodeURL(""); url != "" {
			info.AuthURL = fmt.Sprintf("/auth/oauth/%s", p.Name())
		}
		infos = append(infos, info)
	}
	return infos
}

// HasLocal returns true if a local password provider is configured.
func (r *ProviderRegistry) HasLocal() bool {
	for _, p := range r.providers {
		if p.Type() == ProviderLocal {
			return true
		}
	}
	return false
}
