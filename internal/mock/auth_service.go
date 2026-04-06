package mock

import (
	"context"
	"time"

	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
)

type AuthService struct {
	RegisterFn                         func(ctx context.Context, email, password, displayName string) (*domain.Session, error)
	AllowSignupFn                      func() bool
	LoginWithPasswordFn                func(ctx context.Context, email, password string) (*domain.Session, error)
	GetAuthURLFn                       func(providerName, state string) (string, error)
	HandleOAuthCallbackFn              func(ctx context.Context, providerName, code string) (*domain.Session, error)
	ListProvidersFn                    func() []auth.ProviderInfo
	GetSessionFn                       func(ctx context.Context, sessionID string) (*domain.Session, error)
	DeleteSessionFn                    func(ctx context.Context, sessionID string) error
	CreateAPITokenFn                   func(ctx context.Context, ownerID string, ownerType domain.PrincipalType, name string, expiresAt *time.Time) (string, *domain.APIToken, error)
	ListAPITokensFn                    func(ctx context.Context, ownerID string, ownerType domain.PrincipalType) ([]*domain.APIToken, error)
	RevokeAPITokenFn                   func(ctx context.Context, tokenID string) error
	CreateServiceAccountFn             func(ctx context.Context, name, description string, roles []string) (*domain.ServiceAccount, string, error)
	ListServiceAccountsFn              func(ctx context.Context) ([]*domain.ServiceAccount, error)
	DeleteServiceAccountFn             func(ctx context.Context, id string) error
	RotateServiceAccountSecretFn       func(ctx context.Context, id string) (string, error)
	ResolveAPITokenFn                  func(ctx context.Context, rawToken string) (*domain.Principal, error)
	ResolveServiceAccountCredentialsFn func(ctx context.Context, clientID, clientSecret string) (*domain.Principal, error)
	ResolveSessionFn                   func(ctx context.Context, sessionID string) (*domain.Principal, error)
}

func (m *AuthService) Register(ctx context.Context, email, password, displayName string) (*domain.Session, error) {
	return m.RegisterFn(ctx, email, password, displayName)
}

func (m *AuthService) AllowSignup() bool {
	return m.AllowSignupFn()
}

func (m *AuthService) LoginWithPassword(ctx context.Context, email, password string) (*domain.Session, error) {
	return m.LoginWithPasswordFn(ctx, email, password)
}

func (m *AuthService) GetAuthURL(providerName, state string) (string, error) {
	return m.GetAuthURLFn(providerName, state)
}

func (m *AuthService) HandleOAuthCallback(ctx context.Context, providerName, code string) (*domain.Session, error) {
	return m.HandleOAuthCallbackFn(ctx, providerName, code)
}

func (m *AuthService) ListProviders() []auth.ProviderInfo {
	return m.ListProvidersFn()
}

func (m *AuthService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	return m.GetSessionFn(ctx, sessionID)
}

func (m *AuthService) DeleteSession(ctx context.Context, sessionID string) error {
	return m.DeleteSessionFn(ctx, sessionID)
}

func (m *AuthService) CreateAPIToken(ctx context.Context, ownerID string, ownerType domain.PrincipalType, name string, expiresAt *time.Time) (string, *domain.APIToken, error) {
	return m.CreateAPITokenFn(ctx, ownerID, ownerType, name, expiresAt)
}

func (m *AuthService) ListAPITokens(ctx context.Context, ownerID string, ownerType domain.PrincipalType) ([]*domain.APIToken, error) {
	return m.ListAPITokensFn(ctx, ownerID, ownerType)
}

func (m *AuthService) RevokeAPIToken(ctx context.Context, tokenID string) error {
	return m.RevokeAPITokenFn(ctx, tokenID)
}

func (m *AuthService) CreateServiceAccount(ctx context.Context, name, description string, roles []string) (*domain.ServiceAccount, string, error) {
	return m.CreateServiceAccountFn(ctx, name, description, roles)
}

func (m *AuthService) ListServiceAccounts(ctx context.Context) ([]*domain.ServiceAccount, error) {
	return m.ListServiceAccountsFn(ctx)
}

func (m *AuthService) DeleteServiceAccount(ctx context.Context, id string) error {
	return m.DeleteServiceAccountFn(ctx, id)
}

func (m *AuthService) RotateServiceAccountSecret(ctx context.Context, id string) (string, error) {
	return m.RotateServiceAccountSecretFn(ctx, id)
}

func (m *AuthService) ResolveAPIToken(ctx context.Context, rawToken string) (*domain.Principal, error) {
	return m.ResolveAPITokenFn(ctx, rawToken)
}

func (m *AuthService) ResolveServiceAccountCredentials(ctx context.Context, clientID, clientSecret string) (*domain.Principal, error) {
	return m.ResolveServiceAccountCredentialsFn(ctx, clientID, clientSecret)
}

func (m *AuthService) ResolveSession(ctx context.Context, sessionID string) (*domain.Principal, error) {
	return m.ResolveSessionFn(ctx, sessionID)
}
