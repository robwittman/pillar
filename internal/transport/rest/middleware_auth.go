package rest

import (
	"net/http"
	"strings"

	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

const sessionCookieName = "pillar_session"

// Authenticator is middleware that resolves credentials to a Principal.
// It checks in order: Bearer token, Basic auth, session cookie.
// Returns 401 if no valid credentials are found.
// When the credential source provides an OrgContext (e.g. org-scoped token),
// it is also injected into the request context.
func Authenticator(authSvc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, oc, ok := resolveCredentials(r, authSvc)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
				return
			}
			ctx := auth.ContextWithPrincipal(r.Context(), principal)
			if oc != nil {
				ctx = auth.ContextWithOrg(ctx, oc)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthenticator resolves credentials if present but does not reject unauthenticated requests.
func OptionalAuthenticator(authSvc service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if principal, oc, ok := resolveCredentials(r, authSvc); ok {
				ctx := auth.ContextWithPrincipal(r.Context(), principal)
				if oc != nil {
					ctx = auth.ContextWithOrg(ctx, oc)
				}
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func resolveCredentials(r *http.Request, authSvc service.AuthService) (*domain.Principal, *domain.OrgContext, bool) {
	ctx := r.Context()

	// 1. Bearer token
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if p, oc, err := authSvc.ResolveAPIToken(ctx, token); err == nil {
			return p, oc, true
		}
	}

	// 2. Basic auth (service account credentials)
	if clientID, clientSecret, ok := r.BasicAuth(); ok {
		if p, oc, err := authSvc.ResolveServiceAccountCredentials(ctx, clientID, clientSecret); err == nil {
			return p, oc, true
		}
	}

	// 3. Session cookie (no org context — sessions are user-scoped, org comes from OrgResolver middleware)
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if p, err := authSvc.ResolveSession(ctx, cookie.Value); err == nil {
			return p, nil, true
		}
	}

	return nil, nil, false
}
