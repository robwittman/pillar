package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type OrgService struct {
	CreateFn           func(ctx context.Context, name, slug string) (*domain.Organization, error)
	GetFn              func(ctx context.Context, id string) (*domain.Organization, error)
	ListByUserFn       func(ctx context.Context, userID string) ([]*domain.Organization, error)
	UpdateFn           func(ctx context.Context, id, name, slug string) (*domain.Organization, error)
	DeleteFn           func(ctx context.Context, id string) error
	AddMemberFn        func(ctx context.Context, orgID, userID string, role domain.OrgRole) (*domain.Membership, error)
	RemoveMemberFn     func(ctx context.Context, orgID, userID string) error
	UpdateMemberRoleFn func(ctx context.Context, orgID, userID string, role domain.OrgRole) error
	ListMembersFn      func(ctx context.Context, orgID string) ([]*domain.Membership, error)
	CreateTeamFn       func(ctx context.Context, orgID, name string) (*domain.Team, error)
	ListTeamsFn        func(ctx context.Context, orgID string) ([]*domain.Team, error)
	DeleteTeamFn       func(ctx context.Context, teamID string) error
	AddTeamMemberFn    func(ctx context.Context, teamID, userID string) error
	RemoveTeamMemberFn func(ctx context.Context, teamID, userID string) error
	ListTeamMembersFn  func(ctx context.Context, teamID string) ([]*domain.TeamMembership, error)
}

func (m *OrgService) Create(ctx context.Context, name, slug string) (*domain.Organization, error) {
	return m.CreateFn(ctx, name, slug)
}

func (m *OrgService) Get(ctx context.Context, id string) (*domain.Organization, error) {
	return m.GetFn(ctx, id)
}

func (m *OrgService) ListByUser(ctx context.Context, userID string) ([]*domain.Organization, error) {
	return m.ListByUserFn(ctx, userID)
}

func (m *OrgService) Update(ctx context.Context, id, name, slug string) (*domain.Organization, error) {
	return m.UpdateFn(ctx, id, name, slug)
}

func (m *OrgService) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *OrgService) AddMember(ctx context.Context, orgID, userID string, role domain.OrgRole) (*domain.Membership, error) {
	return m.AddMemberFn(ctx, orgID, userID, role)
}

func (m *OrgService) RemoveMember(ctx context.Context, orgID, userID string) error {
	return m.RemoveMemberFn(ctx, orgID, userID)
}

func (m *OrgService) UpdateMemberRole(ctx context.Context, orgID, userID string, role domain.OrgRole) error {
	return m.UpdateMemberRoleFn(ctx, orgID, userID, role)
}

func (m *OrgService) ListMembers(ctx context.Context, orgID string) ([]*domain.Membership, error) {
	return m.ListMembersFn(ctx, orgID)
}

func (m *OrgService) CreateTeam(ctx context.Context, orgID, name string) (*domain.Team, error) {
	return m.CreateTeamFn(ctx, orgID, name)
}

func (m *OrgService) ListTeams(ctx context.Context, orgID string) ([]*domain.Team, error) {
	return m.ListTeamsFn(ctx, orgID)
}

func (m *OrgService) DeleteTeam(ctx context.Context, teamID string) error {
	return m.DeleteTeamFn(ctx, teamID)
}

func (m *OrgService) AddTeamMember(ctx context.Context, teamID, userID string) error {
	return m.AddTeamMemberFn(ctx, teamID, userID)
}

func (m *OrgService) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	return m.RemoveTeamMemberFn(ctx, teamID, userID)
}

func (m *OrgService) ListTeamMembers(ctx context.Context, teamID string) ([]*domain.TeamMembership, error) {
	return m.ListTeamMembersFn(ctx, teamID)
}
