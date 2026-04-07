package domain

import (
	"context"
	"time"
)

// OrgRole defines the role a user has within an organization.
type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
	OrgRoleViewer OrgRole = "viewer"
)

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Personal  bool      `json:"personal"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	Get(ctx context.Context, id string) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	GetPersonalOrg(ctx context.Context, ownerID string) (*Organization, error)
	ListByUser(ctx context.Context, userID string) ([]*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id string) error
}

type Membership struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	UserID    string    `json:"user_id"`
	Role      OrgRole   `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MembershipRepository interface {
	Create(ctx context.Context, m *Membership) error
	Get(ctx context.Context, id string) (*Membership, error)
	GetByOrgAndUser(ctx context.Context, orgID, userID string) (*Membership, error)
	ListByOrg(ctx context.Context, orgID string) ([]*Membership, error)
	ListByUser(ctx context.Context, userID string) ([]*Membership, error)
	Update(ctx context.Context, m *Membership) error
	Delete(ctx context.Context, id string) error
}

type Team struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TeamRepository interface {
	Create(ctx context.Context, t *Team) error
	Get(ctx context.Context, id string) (*Team, error)
	ListByOrg(ctx context.Context, orgID string) ([]*Team, error)
	Update(ctx context.Context, t *Team) error
	Delete(ctx context.Context, id string) error
}

type TeamMembership struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMembershipRepository interface {
	Add(ctx context.Context, tm *TeamMembership) error
	Remove(ctx context.Context, teamID, userID string) error
	ListByTeam(ctx context.Context, teamID string) ([]*TeamMembership, error)
	ListByUser(ctx context.Context, userID string) ([]*TeamMembership, error)
}

// OrgContext carries the resolved organization scope for the current request.
type OrgContext struct {
	OrgID   string  `json:"org_id"`
	OrgSlug string  `json:"org_slug"`
	OrgRole OrgRole `json:"org_role"`
}
