package rest

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

type AuthHandlers struct {
	authSvc service.AuthService
	logger  *slog.Logger
}

func NewAuthHandlers(authSvc service.AuthService, logger *slog.Logger) *AuthHandlers {
	return &AuthHandlers{authSvc: authSvc, logger: logger}
}

// ListProviders returns configured auth providers for the login page.
func (h *AuthHandlers) ListProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"providers":    h.authSvc.ListProviders(),
		"allow_signup": h.authSvc.AllowSignup(),
	})
}

// Login handles local username/password authentication.
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	session, err := h.authSvc.LoginWithPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	setSessionCookie(w, session)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Register creates a new local user account.
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	session, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			writeError(w, http.StatusConflict, "a user with that email already exists")
			return
		}
		writeError(w, http.StatusBadRequest, "registration failed")
		return
	}

	setSessionCookie(w, session)
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// OAuthRedirect initiates an OAuth/OIDC flow by redirecting to the provider.
func (h *AuthHandlers) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	state, err := generateOAuthState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}

	// Store state in a short-lived cookie for CSRF verification on callback.
	http.SetCookie(w, &http.Cookie{
		Name:     "pillar_oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url, err := h.authSvc.GetAuthURL(providerName, state)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unknown provider")
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

// OAuthCallback handles the redirect back from an OAuth/OIDC provider.
func (h *AuthHandlers) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")

	// Verify CSRF state.
	stateCookie, err := r.Cookie("pillar_oauth_state")
	if err != nil || stateCookie.Value == "" {
		writeError(w, http.StatusBadRequest, "missing oauth state")
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		writeError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	// Clear state cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "pillar_oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	session, err := h.authSvc.HandleOAuthCallback(r.Context(), providerName, code)
	if err != nil {
		h.logger.Warn("OAuth callback failed", "provider", providerName, "error", err)
		writeError(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	setSessionCookie(w, session)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Logout deletes the session and clears the cookie.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		_ = h.authSvc.DeleteSession(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me returns the current authenticated user.
func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, principal)
}

// --- API Token handlers ---

type createTokenRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CreateToken creates a new API token for the authenticated caller.
func (h *AuthHandlers) CreateToken(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req createTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	rawToken, meta, err := h.authSvc.CreateAPIToken(r.Context(), principal.ID, principal.Type, req.Name, req.ExpiresAt)
	if err != nil {
		h.logger.Error("failed to create token", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": rawToken,
		"meta":  meta,
	})
}

// ListTokens lists API tokens for the authenticated caller.
func (h *AuthHandlers) ListTokens(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	tokens, err := h.authSvc.ListAPITokens(r.Context(), principal.ID, principal.Type)
	if err != nil {
		h.logger.Error("failed to list tokens", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list tokens")
		return
	}
	if tokens == nil {
		tokens = []*domain.APIToken{}
	}

	writeJSON(w, http.StatusOK, tokens)
}

// RevokeToken deletes an API token.
func (h *AuthHandlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	tokenID := chi.URLParam(r, "tokenID")
	if err := h.authSvc.RevokeAPIToken(r.Context(), tokenID); err != nil {
		writeError(w, http.StatusNotFound, "token not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Service Account handlers ---

type createServiceAccountRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}

// CreateServiceAccount creates a new service account.
func (h *AuthHandlers) CreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	var req createServiceAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	sa, secret, err := h.authSvc.CreateServiceAccount(r.Context(), req.Name, req.Description, req.Roles)
	if err != nil {
		h.logger.Error("failed to create service account", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create service account")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"service_account": sa,
		"client_id":       sa.ID,
		"client_secret":   secret,
	})
}

// ListServiceAccounts lists all service accounts.
func (h *AuthHandlers) ListServiceAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.authSvc.ListServiceAccounts(r.Context())
	if err != nil {
		h.logger.Error("failed to list service accounts", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list service accounts")
		return
	}
	if accounts == nil {
		accounts = []*domain.ServiceAccount{}
	}

	writeJSON(w, http.StatusOK, accounts)
}

// DeleteServiceAccount deletes a service account.
func (h *AuthHandlers) DeleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.authSvc.DeleteServiceAccount(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "service account not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RotateServiceAccountSecret rotates the secret for a service account.
func (h *AuthHandlers) RotateServiceAccountSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	newSecret, err := h.authSvc.RotateServiceAccountSecret(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "service account not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"client_id":     id,
		"client_secret": newSecret,
	})
}

// --- Helpers ---

func setSessionCookie(w http.ResponseWriter, session *domain.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ReconcilePersonalOrgs finds users without personal orgs and creates them.
func (h *AuthHandlers) ReconcilePersonalOrgs(w http.ResponseWriter, r *http.Request) {
	// Check that caller is an admin.
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	isAdmin := false
	for _, role := range principal.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		writeError(w, http.StatusForbidden, "admin role required")
		return
	}

	result, err := h.authSvc.ReconcilePersonalOrgs(r.Context())
	if err != nil {
		h.logger.Error("reconciliation failed", "error", err)
		writeError(w, http.StatusInternalServerError, "reconciliation failed")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
