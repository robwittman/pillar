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

// Bootstrap checks if any users exist. If not, it creates an admin user.
// If adminPassword is empty, a random password is generated.
// If adminEmail is empty, it defaults to "admin@pillar.local".
func Bootstrap(ctx context.Context, userRepo domain.UserRepository, adminEmail, adminPassword string, logger *slog.Logger) (*BootstrapResult, error) {
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

	result := &BootstrapResult{
		Created: true,
		Email:   adminEmail,
	}
	if generated {
		result.Password = adminPassword
	}
	return result, nil
}
