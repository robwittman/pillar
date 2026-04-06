package auth

import (
	"context"
	"fmt"

	"github.com/robwittman/pillar/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// LocalProvider authenticates users via email/password against the users table.
type LocalProvider struct {
	userRepo domain.UserRepository
}

func NewLocalProvider(userRepo domain.UserRepository) *LocalProvider {
	return &LocalProvider{userRepo: userRepo}
}

func (p *LocalProvider) Type() ProviderType { return ProviderLocal }
func (p *LocalProvider) Name() string       { return "local" }

func (p *LocalProvider) AuthCodeURL(_ string) string { return "" }

func (p *LocalProvider) ExchangeCode(_ context.Context, _ string) (*ExternalIdentity, error) {
	return nil, fmt.Errorf("local provider does not support OAuth code exchange")
}

func (p *LocalProvider) ValidateCredentials(ctx context.Context, email, password string) (*ExternalIdentity, error) {
	user, err := p.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user.Disabled {
		return nil, domain.ErrInvalidCredentials
	}
	if user.PasswordHash == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return &ExternalIdentity{
		Provider:    "local",
		SubjectID:   user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}, nil
}

// HashPassword generates a bcrypt hash for a plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}
