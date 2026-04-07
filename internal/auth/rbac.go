package auth

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

// RequireOrgRole returns an error if the current org context role is not one of the required roles.
func RequireOrgRole(ctx context.Context, required ...domain.OrgRole) error {
	oc, ok := OrgFromContext(ctx)
	if !ok {
		return domain.ErrOrgContextRequired
	}
	for _, r := range required {
		if oc.OrgRole == r {
			return nil
		}
	}
	return domain.ErrNotAuthorized
}

// CanManageResources returns true if the role can create/update/delete resources.
func CanManageResources(role domain.OrgRole) bool {
	return role == domain.OrgRoleOwner || role == domain.OrgRoleAdmin || role == domain.OrgRoleMember
}

// CanManageOrg returns true if the role can manage org settings and members.
func CanManageOrg(role domain.OrgRole) bool {
	return role == domain.OrgRoleOwner || role == domain.OrgRoleAdmin
}
