package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/domain"
)

type OrgService interface {
	// Organizations
	Create(ctx context.Context, name, slug string) (*domain.Organization, error)
	Get(ctx context.Context, id string) (*domain.Organization, error)
	ListByUser(ctx context.Context, userID string) ([]*domain.Organization, error)
	Update(ctx context.Context, id, name, slug string) (*domain.Organization, error)
	Delete(ctx context.Context, id string) error

	// Memberships
	AddMember(ctx context.Context, orgID, userID string, role domain.OrgRole) (*domain.Membership, error)
	RemoveMember(ctx context.Context, orgID, userID string) error
	UpdateMemberRole(ctx context.Context, orgID, userID string, role domain.OrgRole) error
	ListMembers(ctx context.Context, orgID string) ([]*domain.Membership, error)

	// Teams
	CreateTeam(ctx context.Context, orgID, name string) (*domain.Team, error)
	ListTeams(ctx context.Context, orgID string) ([]*domain.Team, error)
	DeleteTeam(ctx context.Context, teamID string) error
	AddTeamMember(ctx context.Context, teamID, userID string) error
	RemoveTeamMember(ctx context.Context, teamID, userID string) error
	ListTeamMembers(ctx context.Context, teamID string) ([]*domain.TeamMembership, error)
}

type orgService struct {
	orgRepo        domain.OrganizationRepository
	membershipRepo domain.MembershipRepository
	teamRepo       domain.TeamRepository
	teamMemberRepo domain.TeamMembershipRepository
	logger         *slog.Logger
}

func NewOrgService(
	orgRepo domain.OrganizationRepository,
	membershipRepo domain.MembershipRepository,
	teamRepo domain.TeamRepository,
	teamMemberRepo domain.TeamMembershipRepository,
	logger *slog.Logger,
) OrgService {
	return &orgService{
		orgRepo:        orgRepo,
		membershipRepo: membershipRepo,
		teamRepo:       teamRepo,
		teamMemberRepo: teamMemberRepo,
		logger:         logger,
	}
}

// --- Organizations ---

func (s *orgService) Create(ctx context.Context, name, slug string) (*domain.Organization, error) {
	principal, err := requirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	org := &domain.Organization{
		ID:      uuid.New().String(),
		Name:    name,
		Slug:    slug,
		OwnerID: principal.ID,
	}
	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	// Creator becomes owner.
	membership := &domain.Membership{
		ID:     uuid.New().String(),
		OrgID:  org.ID,
		UserID: principal.ID,
		Role:   domain.OrgRoleOwner,
	}
	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		return nil, err
	}

	s.logger.Info("organization created", "org_id", org.ID, "name", name, "owner", principal.ID)
	return org, nil
}

func (s *orgService) Get(ctx context.Context, id string) (*domain.Organization, error) {
	return s.orgRepo.Get(ctx, id)
}

func (s *orgService) ListByUser(ctx context.Context, userID string) ([]*domain.Organization, error) {
	return s.orgRepo.ListByUser(ctx, userID)
}

func (s *orgService) Update(ctx context.Context, id, name, slug string) (*domain.Organization, error) {
	org, err := s.orgRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if org.Personal {
		return nil, fmt.Errorf("cannot rename a personal organization")
	}

	if name != "" {
		org.Name = name
	}
	if slug != "" {
		org.Slug = slug
	}
	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *orgService) Delete(ctx context.Context, id string) error {
	org, err := s.orgRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if org.Personal {
		return fmt.Errorf("cannot delete a personal organization")
	}
	return s.orgRepo.Delete(ctx, id)
}

// --- Memberships ---

func (s *orgService) AddMember(ctx context.Context, orgID, userID string, role domain.OrgRole) (*domain.Membership, error) {
	// Verify org exists.
	if _, err := s.orgRepo.Get(ctx, orgID); err != nil {
		return nil, err
	}

	m := &domain.Membership{
		ID:     uuid.New().String(),
		OrgID:  orgID,
		UserID: userID,
		Role:   role,
	}
	if err := s.membershipRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *orgService) RemoveMember(ctx context.Context, orgID, userID string) error {
	m, err := s.membershipRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		return err
	}

	// Prevent removing the last owner.
	if m.Role == domain.OrgRoleOwner {
		members, err := s.membershipRepo.ListByOrg(ctx, orgID)
		if err != nil {
			return err
		}
		ownerCount := 0
		for _, mem := range members {
			if mem.Role == domain.OrgRoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return fmt.Errorf("cannot remove the last owner")
		}
	}

	return s.membershipRepo.Delete(ctx, m.ID)
}

func (s *orgService) UpdateMemberRole(ctx context.Context, orgID, userID string, role domain.OrgRole) error {
	m, err := s.membershipRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		return err
	}

	// Prevent demoting the last owner.
	if m.Role == domain.OrgRoleOwner && role != domain.OrgRoleOwner {
		members, err := s.membershipRepo.ListByOrg(ctx, orgID)
		if err != nil {
			return err
		}
		ownerCount := 0
		for _, mem := range members {
			if mem.Role == domain.OrgRoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return fmt.Errorf("cannot demote the last owner")
		}
	}

	m.Role = role
	return s.membershipRepo.Update(ctx, m)
}

func (s *orgService) ListMembers(ctx context.Context, orgID string) ([]*domain.Membership, error) {
	return s.membershipRepo.ListByOrg(ctx, orgID)
}

// --- Teams ---

func (s *orgService) CreateTeam(ctx context.Context, orgID, name string) (*domain.Team, error) {
	t := &domain.Team{
		ID:    uuid.New().String(),
		OrgID: orgID,
		Name:  name,
	}
	if err := s.teamRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *orgService) ListTeams(ctx context.Context, orgID string) ([]*domain.Team, error) {
	return s.teamRepo.ListByOrg(ctx, orgID)
}

func (s *orgService) DeleteTeam(ctx context.Context, teamID string) error {
	return s.teamRepo.Delete(ctx, teamID)
}

func (s *orgService) AddTeamMember(ctx context.Context, teamID, userID string) error {
	tm := &domain.TeamMembership{
		ID:     uuid.New().String(),
		TeamID: teamID,
		UserID: userID,
	}
	return s.teamMemberRepo.Add(ctx, tm)
}

func (s *orgService) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	return s.teamMemberRepo.Remove(ctx, teamID, userID)
}

func (s *orgService) ListTeamMembers(ctx context.Context, teamID string) ([]*domain.TeamMembership, error) {
	return s.teamMemberRepo.ListByTeam(ctx, teamID)
}

func requirePrincipal(ctx context.Context) (*domain.Principal, error) {
	p, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return nil, domain.ErrAuthRequired
	}
	return p, nil
}
