package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

// BootstrapResult holds the outcome of an admin bootstrap attempt.
type BootstrapResult struct {
	Created  bool
	Email    string
	Password string // only set when auto-generated
}

// Bootstrap checks if any users exist. If not, it creates an admin user
// and their personal organization. If adminPassword is empty, a random
// password is generated. If adminEmail is empty, it defaults to "admin@pillar.local".
func Bootstrap(ctx context.Context, userRepo domain.UserRepository, orgRepo domain.OrganizationRepository, membershipRepo domain.MembershipRepository, adminEmail, adminPassword string, logger *slog.Logger) (*BootstrapResult, error) {
	users, err := userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking existing users: %w", err)
	}
	if len(users) > 0 {
		return &BootstrapResult{Created: false}, nil
	}

	if adminEmail == "" {
		adminEmail = "admin@pillar.local"
	}

	generated := false
	if adminPassword == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("generating password: %w", err)
		}
		adminPassword = base64.RawURLEncoding.EncodeToString(b)
		generated = true
	}

	hash, err := HashPassword(adminPassword)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        adminEmail,
		DisplayName:  "Admin",
		PasswordHash: hash,
		Provider:     "local",
		Roles:        []string{"admin"},
	}
	if err := userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating admin user: %w", err)
	}

	logger.Info("bootstrap: created admin user", "email", adminEmail, "user_id", user.ID)

	// Create personal org for the admin user.
	if orgRepo != nil && membershipRepo != nil {
		if err := CreatePersonalOrg(ctx, orgRepo, membershipRepo, user); err != nil {
			logger.Error("bootstrap: failed to create personal org", "user_id", user.ID, "error", err)
		}
	}

	result := &BootstrapResult{
		Created: true,
		Email:   adminEmail,
	}
	if generated {
		result.Password = adminPassword
	}
	return result, nil
}

// CreatePersonalOrg creates a personal organization and owner membership for a user.
// Safe to call multiple times — returns nil if the personal org already exists.
func CreatePersonalOrg(ctx context.Context, orgRepo domain.OrganizationRepository, membershipRepo domain.MembershipRepository, user *domain.User) error {
	// Check if personal org already exists.
	if _, err := orgRepo.GetPersonalOrg(ctx, user.ID); err == nil {
		return nil
	}

	name := user.DisplayName
	if name == "" {
		name = user.Email
	}

	org := &domain.Organization{
		ID:       uuid.New().String(),
		Name:     name + "'s Workspace",
		Slug:     "personal-" + user.ID[:8],
		Personal: true,
		OwnerID:  user.ID,
	}
	if err := orgRepo.Create(ctx, org); err != nil {
		return fmt.Errorf("creating personal org: %w", err)
	}

	membership := &domain.Membership{
		ID:     uuid.New().String(),
		OrgID:  org.ID,
		UserID: user.ID,
		Role:   domain.OrgRoleOwner,
	}
	if err := membershipRepo.Create(ctx, membership); err != nil {
		return fmt.Errorf("creating owner membership: %w", err)
	}

	return nil
}
