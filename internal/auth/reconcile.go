package auth

import (
	"context"
	"log/slog"

	"github.com/robwittman/pillar/internal/domain"
)

// ReconcileResult holds the outcome of a personal org reconciliation run.
type ReconcileResult struct {
	Checked int
	Created int
	Errors  int
}

// ReconcilePersonalOrgs finds all users without a personal org and creates one.
// This is safe to run repeatedly — it's idempotent.
func ReconcilePersonalOrgs(ctx context.Context, userRepo domain.UserRepository, orgRepo domain.OrganizationRepository, membershipRepo domain.MembershipRepository, logger *slog.Logger) (*ReconcileResult, error) {
	users, err := userRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := &ReconcileResult{Checked: len(users)}

	for _, user := range users {
		_, err := orgRepo.GetPersonalOrg(ctx, user.ID)
		if err == nil {
			continue // already has personal org
		}

		logger.Info("creating missing personal org", "user_id", user.ID, "email", user.Email)
		if err := CreatePersonalOrg(ctx, orgRepo, membershipRepo, user); err != nil {
			logger.Error("failed to create personal org", "user_id", user.ID, "error", err)
			result.Errors++
			continue
		}
		result.Created++
	}

	return result, nil
}
