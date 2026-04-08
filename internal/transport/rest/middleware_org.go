package rest

import (
	"net/http"

	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
)

// OrgResolver is middleware that resolves the organization context for the request.
// It checks the X-Org-ID header first. If absent, it auto-selects the user's org
// when they belong to exactly one. Injects OrgContext into the request context.
func OrgResolver(membershipRepo domain.MembershipRepository, orgRepo domain.OrganizationRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := auth.PrincipalFromContext(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			// Skip if org context was already set by credential resolution
			// (e.g. org-scoped API token or service account Basic auth).
			if _, hasOrg := auth.OrgFromContext(r.Context()); hasOrg {
				next.ServeHTTP(w, r)
				return
			}

			orgHeader := r.Header.Get("X-Org-ID")

			if orgHeader == "" {
				// Auto-select: if user has exactly one org, use it.
				memberships, err := membershipRepo.ListByUser(r.Context(), principal.ID)
				if err != nil || len(memberships) == 0 {
					writeError(w, http.StatusBadRequest, "organization context required")
					return
				}
				if len(memberships) == 1 {
					org, err := orgRepo.Get(r.Context(), memberships[0].OrgID)
					if err != nil {
						writeError(w, http.StatusInternalServerError, "failed to resolve organization")
						return
					}
					ctx := auth.ContextWithOrg(r.Context(), &domain.OrgContext{
						OrgID:   org.ID,
						OrgSlug: org.Slug,
						OrgRole: domain.OrgRole(memberships[0].Role),
					})
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				writeError(w, http.StatusBadRequest, "X-Org-ID header required when user belongs to multiple organizations")
				return
			}

			// Resolve org by ID or slug.
			org, err := orgRepo.Get(r.Context(), orgHeader)
			if err != nil {
				org, err = orgRepo.GetBySlug(r.Context(), orgHeader)
			}
			if err != nil {
				writeError(w, http.StatusNotFound, "organization not found")
				return
			}

			// Verify the principal is a member.
			membership, err := membershipRepo.GetByOrgAndUser(r.Context(), org.ID, principal.ID)
			if err != nil {
				writeError(w, http.StatusForbidden, "not a member of this organization")
				return
			}

			ctx := auth.ContextWithOrg(r.Context(), &domain.OrgContext{
				OrgID:   org.ID,
				OrgSlug: org.Slug,
				OrgRole: domain.OrgRole(membership.Role),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
