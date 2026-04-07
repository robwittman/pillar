package service

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines operations for authentication and credential management.
type AuthService interface {
	// Register a new local user (only when allow_signup is true).
	Register(ctx context.Context, email, password, displayName string) (*domain.Session, error)

	// Whether self-registration is allowed.
	AllowSignup() bool

	// Password login (local provider only).
	LoginWithPassword(ctx context.Context, email, password string) (*domain.Session, error)

	// OAuth/OIDC: get redirect URL for a provider.
	GetAuthURL(providerName, state string) (string, error)

	// OAuth/OIDC: exchange authorization code for a session.
	HandleOAuthCallback(ctx context.Context, providerName, code string) (*domain.Session, error)

	// List configured auth providers (for login page).
	ListProviders() []auth.ProviderInfo

	// Session management.
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error

	// API token management.
	CreateAPIToken(ctx context.Context, ownerID string, ownerType domain.PrincipalType, name string, expiresAt *time.Time) (rawToken string, meta *domain.APIToken, err error)
	ListAPITokens(ctx context.Context, ownerID string, ownerType domain.PrincipalType) ([]*domain.APIToken, error)
	RevokeAPIToken(ctx context.Context, tokenID string) error

	// Service account management.
	CreateServiceAccount(ctx context.Context, name, description string, roles []string) (sa *domain.ServiceAccount, secret string, err error)
	ListServiceAccounts(ctx context.Context) ([]*domain.ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, id string) error
	RotateServiceAccountSecret(ctx context.Context, id string) (newSecret string, err error)

	// Admin operations.
	ReconcilePersonalOrgs(ctx context.Context) (*auth.ReconcileResult, error)

	// Credential resolution (used by middleware).
	// Resolve methods return a Principal and optionally an OrgContext (nil when no org is bound).
	ResolveAPIToken(ctx context.Context, rawToken string) (*domain.Principal, *domain.OrgContext, error)
	ResolveServiceAccountCredentials(ctx context.Context, clientID, clientSecret string) (*domain.Principal, *domain.OrgContext, error)
	ResolveSession(ctx context.Context, sessionID string) (*domain.Principal, error)
}

type authService struct {
	userRepo       domain.UserRepository
	saRepo         domain.ServiceAccountRepository
	tokenRepo      domain.APITokenRepository
	sessionStore   domain.SessionStore
	orgRepo        domain.OrganizationRepository
	membershipRepo domain.MembershipRepository
	providers      *auth.ProviderRegistry
	sessionTTL     time.Duration
	allowSignup    bool
	logger         *slog.Logger
}

func NewAuthService(
	userRepo domain.UserRepository,
	saRepo domain.ServiceAccountRepository,
	tokenRepo domain.APITokenRepository,
	sessionStore domain.SessionStore,
	orgRepo domain.OrganizationRepository,
	membershipRepo domain.MembershipRepository,
	providers *auth.ProviderRegistry,
	sessionTTL time.Duration,
	allowSignup bool,
	logger *slog.Logger,
) AuthService {
	return &authService{
		userRepo:       userRepo,
		saRepo:         saRepo,
		tokenRepo:      tokenRepo,
		sessionStore:   sessionStore,
		orgRepo:        orgRepo,
		membershipRepo: membershipRepo,
		providers:      providers,
		sessionTTL:     sessionTTL,
		allowSignup:    allowSignup,
		logger:         logger,
	}
}

func (s *authService) AllowSignup() bool {
	return s.allowSignup && s.providers.HasLocal()
}

func (s *authService) Register(ctx context.Context, email, password, displayName string) (*domain.Session, error) {
	if !s.allowSignup {
		return nil, domain.ErrInvalidCredentials
	}
	if !s.providers.HasLocal() {
		return nil, domain.ErrInvalidCredentials
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	if displayName == "" {
		displayName = email
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: hash,
		Provider:     "local",
		Roles:        []string{"member"},
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	if err := s.createPersonalOrg(ctx, user); err != nil {
		s.logger.Error("failed to create personal org for new user", "user_id", user.ID, "error", err)
	}

	s.logger.Info("new user registered", "user_id", user.ID, "email", email)
	return s.createSession(ctx, user.ID)
}

func (s *authService) LoginWithPassword(ctx context.Context, email, password string) (*domain.Session, error) {
	// Find a local provider.
	var localProvider auth.IdentityProvider
	for _, info := range s.providers.List() {
		if info.Type == auth.ProviderLocal {
			p, _ := s.providers.Get(info.Name)
			localProvider = p
			break
		}
	}
	if localProvider == nil {
		return nil, domain.ErrInvalidCredentials
	}

	identity, err := localProvider.ValidateCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// For local auth, SubjectID is the user ID.
	user, err := s.userRepo.Get(ctx, identity.SubjectID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.createSession(ctx, user.ID)
}

func (s *authService) GetAuthURL(providerName, state string) (string, error) {
	p, ok := s.providers.Get(providerName)
	if !ok {
		return "", domain.ErrInvalidCredentials
	}

	url := p.AuthCodeURL(state)
	if url == "" {
		return "", domain.ErrInvalidCredentials
	}
	return url, nil
}

func (s *authService) HandleOAuthCallback(ctx context.Context, providerName, code string) (*domain.Session, error) {
	p, ok := s.providers.Get(providerName)
	if !ok {
		return nil, domain.ErrInvalidCredentials
	}

	identity, err := p.ExchangeCode(ctx, code)
	if err != nil {
		s.logger.Warn("OAuth code exchange failed", "provider", providerName, "error", err)
		return nil, domain.ErrInvalidCredentials
	}

	// Look up or create user by provider + subject ID.
	user, err := s.userRepo.GetByProviderSub(ctx, identity.Provider, identity.SubjectID)
	if err != nil {
		if err != domain.ErrUserNotFound {
			return nil, err
		}

		// User not found — create if signup is allowed.
		if !s.allowSignup {
			s.logger.Warn("signup not allowed for new OAuth user", "provider", providerName, "email", identity.Email)
			return nil, domain.ErrInvalidCredentials
		}

		user = &domain.User{
			ID:            uuid.New().String(),
			Email:         identity.Email,
			DisplayName:   identity.DisplayName,
			Provider:      identity.Provider,
			ProviderSubID: identity.SubjectID,
			Roles:         []string{"member"},
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
		if err := s.createPersonalOrg(ctx, user); err != nil {
			s.logger.Error("failed to create personal org for OAuth user", "user_id", user.ID, "error", err)
		}
		s.logger.Info("created new user via OAuth", "provider", providerName, "user_id", user.ID, "email", user.Email)
	}

	if user.Disabled {
		return nil, domain.ErrInvalidCredentials
	}

	return s.createSession(ctx, user.ID)
}

func (s *authService) ListProviders() []auth.ProviderInfo {
	return s.providers.List()
}

func (s *authService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	return s.sessionStore.Get(ctx, sessionID)
}

func (s *authService) DeleteSession(ctx context.Context, sessionID string) error {
	return s.sessionStore.Delete(ctx, sessionID)
}

// --- API Tokens ---

func (s *authService) CreateAPIToken(ctx context.Context, ownerID string, ownerType domain.PrincipalType, name string, expiresAt *time.Time) (string, *domain.APIToken, error) {
	tokenID := uuid.New().String()

	rawToken, tokenHash, err := auth.GenerateToken(tokenID)
	if err != nil {
		return "", nil, err
	}

	token := &domain.APIToken{
		ID:        tokenID,
		Name:      name,
		TokenHash: tokenHash,
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Scopes:    []string{},
		ExpiresAt: expiresAt,
	}
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return "", nil, err
	}

	return rawToken, token, nil
}

func (s *authService) ListAPITokens(ctx context.Context, ownerID string, ownerType domain.PrincipalType) ([]*domain.APIToken, error) {
	return s.tokenRepo.ListByOwner(ctx, ownerID, ownerType)
}

func (s *authService) RevokeAPIToken(ctx context.Context, tokenID string) error {
	return s.tokenRepo.Delete(ctx, tokenID)
}

// --- Service Accounts ---

func (s *authService) CreateServiceAccount(ctx context.Context, name, description string, roles []string) (*domain.ServiceAccount, string, error) {
	secret, err := auth.GenerateSecret()
	if err != nil {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	if roles == nil {
		roles = []string{"member"}
	}

	sa := &domain.ServiceAccount{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		SecretHash:  string(hash),
		Roles:       roles,
	}
	if err := s.saRepo.Create(ctx, sa); err != nil {
		return nil, "", err
	}

	return sa, secret, nil
}

func (s *authService) ListServiceAccounts(ctx context.Context) ([]*domain.ServiceAccount, error) {
	return s.saRepo.List(ctx)
}

func (s *authService) DeleteServiceAccount(ctx context.Context, id string) error {
	return s.saRepo.Delete(ctx, id)
}

func (s *authService) RotateServiceAccountSecret(ctx context.Context, id string) (string, error) {
	sa, err := s.saRepo.Get(ctx, id)
	if err != nil {
		return "", err
	}

	secret, err := auth.GenerateSecret()
	if err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	sa.SecretHash = string(hash)
	if err := s.saRepo.Update(ctx, sa); err != nil {
		return "", err
	}

	return secret, nil
}

// --- Admin Operations ---

func (s *authService) ReconcilePersonalOrgs(ctx context.Context) (*auth.ReconcileResult, error) {
	if s.orgRepo == nil || s.membershipRepo == nil {
		return &auth.ReconcileResult{}, nil
	}
	return auth.ReconcilePersonalOrgs(ctx, s.userRepo, s.orgRepo, s.membershipRepo, s.logger)
}

// --- Credential Resolution (used by middleware) ---

func (s *authService) ResolveAPIToken(ctx context.Context, rawToken string) (*domain.Principal, *domain.OrgContext, error) {
	tokenHash := auth.HashToken(rawToken)

	token, err := s.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, domain.ErrAuthRequired
	}

	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return nil, nil, domain.ErrTokenExpired
	}

	// Update last-used timestamp asynchronously (best-effort).
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.tokenRepo.UpdateLastUsed(bgCtx, token.ID)
	}()

	var principal *domain.Principal

	// Resolve the owner to build the principal.
	switch token.OwnerType {
	case domain.PrincipalUser:
		user, err := s.userRepo.Get(ctx, token.OwnerID)
		if err != nil {
			return nil, nil, domain.ErrAuthRequired
		}
		if user.Disabled {
			return nil, nil, domain.ErrAuthRequired
		}
		principal = &domain.Principal{
			ID:          user.ID,
			Type:        domain.PrincipalUser,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			TokenID:     token.ID,
			Roles:       user.Roles,
		}

	case domain.PrincipalServiceAccount:
		sa, err := s.saRepo.Get(ctx, token.OwnerID)
		if err != nil {
			return nil, nil, domain.ErrAuthRequired
		}
		if sa.Disabled {
			return nil, nil, domain.ErrAuthRequired
		}
		principal = &domain.Principal{
			ID:          sa.ID,
			Type:        domain.PrincipalServiceAccount,
			DisplayName: sa.Name,
			TokenID:     token.ID,
			Roles:       sa.Roles,
		}

	default:
		return nil, nil, domain.ErrAuthRequired
	}

	// Resolve org context from the token's org_id.
	oc := s.resolveOrgContext(ctx, token.OrgID, principal.ID)
	return principal, oc, nil
}

func (s *authService) ResolveServiceAccountCredentials(ctx context.Context, clientID, clientSecret string) (*domain.Principal, *domain.OrgContext, error) {
	sa, err := s.saRepo.Get(ctx, clientID)
	if err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}
	if sa.Disabled {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(sa.SecretHash), []byte(clientSecret)); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	principal := &domain.Principal{
		ID:          sa.ID,
		Type:        domain.PrincipalServiceAccount,
		DisplayName: sa.Name,
		Roles:       sa.Roles,
	}

	oc := s.resolveOrgContext(ctx, sa.OrgID, sa.ID)
	return principal, oc, nil
}

// resolveOrgContext looks up the org and membership to build an OrgContext.
// Returns nil if orgID is empty or repos aren't configured.
func (s *authService) resolveOrgContext(ctx context.Context, orgID, principalID string) *domain.OrgContext {
	if orgID == "" || s.orgRepo == nil || s.membershipRepo == nil {
		return nil
	}

	org, err := s.orgRepo.Get(ctx, orgID)
	if err != nil {
		return nil
	}

	membership, err := s.membershipRepo.GetByOrgAndUser(ctx, orgID, principalID)
	if err != nil {
		// Service accounts may not have an explicit membership — default to member role.
		return &domain.OrgContext{
			OrgID:   org.ID,
			OrgSlug: org.Slug,
			OrgRole: domain.OrgRoleMember,
		}
	}

	return &domain.OrgContext{
		OrgID:   org.ID,
		OrgSlug: org.Slug,
		OrgRole: domain.OrgRole(membership.Role),
	}
}

func (s *authService) ResolveSession(ctx context.Context, sessionID string) (*domain.Principal, error) {
	session, err := s.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrAuthRequired
	}

	user, err := s.userRepo.Get(ctx, session.UserID)
	if err != nil {
		return nil, domain.ErrAuthRequired
	}
	if user.Disabled {
		return nil, domain.ErrAuthRequired
	}

	return &domain.Principal{
		ID:          user.ID,
		Type:        domain.PrincipalUser,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Roles:       user.Roles,
	}, nil
}

// --- Helpers ---

func (s *authService) createSession(ctx context.Context, userID string) (*domain.Session, error) {
	sessionID, err := auth.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.sessionTTL),
		CreatedAt: time.Now(),
	}
	if err := s.sessionStore.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *authService) createPersonalOrg(ctx context.Context, user *domain.User) error {
	if s.orgRepo == nil || s.membershipRepo == nil {
		return nil // orgs not configured
	}

	// Check if personal org already exists.
	if _, err := s.orgRepo.GetPersonalOrg(ctx, user.ID); err == nil {
		return nil // already exists
	}

	name := user.DisplayName
	if name == "" {
		name = user.Email
	}

	org := &domain.Organization{
		ID:       uuid.New().String(),
		Name:     name + "'s Workspace",
		Slug:     slugFromEmail(user.Email) + "-" + user.ID[:8],
		Personal: true,
		OwnerID:  user.ID,
	}
	if err := s.orgRepo.Create(ctx, org); err != nil {
		return err
	}

	membership := &domain.Membership{
		ID:     uuid.New().String(),
		OrgID:  org.ID,
		UserID: user.ID,
		Role:   domain.OrgRoleOwner,
	}
	return s.membershipRepo.Create(ctx, membership)
}

var slugRegexp = regexp.MustCompile(`[^a-z0-9-]`)

func slugFromEmail(email string) string {
	local := strings.SplitN(email, "@", 2)[0]
	slug := strings.ToLower(local)
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = slugRegexp.ReplaceAllString(slug, "")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "user"
	}
	return slug
}
