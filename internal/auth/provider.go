package auth

import "context"

// ProviderType identifies the kind of identity provider.
type ProviderType string

const (
	ProviderLocal  ProviderType = "local"
	ProviderOIDC   ProviderType = "oidc"
	ProviderGitHub ProviderType = "github"
)

// ExternalIdentity is the user information returned by an identity provider
// after successful authentication.
type ExternalIdentity struct {
	Provider    string
	SubjectID   string
	Email       string
	DisplayName string
}

// IdentityProvider handles authentication via a specific method.
type IdentityProvider interface {
	// Type returns the provider type (local, oidc, github).
	Type() ProviderType

	// Name returns the unique display name for this provider instance.
	Name() string

	// AuthCodeURL returns the OAuth/OIDC authorization redirect URL.
	// Returns empty string for providers that don't use OAuth (e.g. local).
	AuthCodeURL(state string) string

	// ExchangeCode exchanges an OAuth authorization code for user identity.
	// Only applicable to OAuth/OIDC providers.
	ExchangeCode(ctx context.Context, code string) (*ExternalIdentity, error)

	// ValidateCredentials validates a username/password pair.
	// Only applicable to the local provider.
	ValidateCredentials(ctx context.Context, email, password string) (*ExternalIdentity, error)
}
