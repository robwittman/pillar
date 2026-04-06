package domain

import (
	"context"
	"time"
)

// PrincipalType identifies the kind of authenticated caller.
type PrincipalType string

const (
	PrincipalUser           PrincipalType = "user"
	PrincipalServiceAccount PrincipalType = "service_account"
	PrincipalAgent          PrincipalType = "agent"
)

// Principal represents an authenticated caller resolved by middleware.
type Principal struct {
	ID          string        `json:"id"`
	Type        PrincipalType `json:"type"`
	DisplayName string        `json:"display_name"`
	Email       string        `json:"email,omitempty"`
	TokenID     string        `json:"token_id,omitempty"`
	Roles       []string      `json:"roles"`
}

// User represents a human account.
type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	PasswordHash  string    `json:"-"`
	Provider      string    `json:"provider"`
	ProviderSubID string    `json:"provider_sub_id,omitempty"`
	Roles         []string  `json:"roles"`
	Disabled      bool      `json:"disabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Get(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByProviderSub(ctx context.Context, provider, subID string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// ServiceAccount represents a bot / non-human account.
type ServiceAccount struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SecretHash  string    `json:"-"`
	Roles       []string  `json:"roles"`
	Disabled    bool      `json:"disabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ServiceAccountRepository interface {
	Create(ctx context.Context, sa *ServiceAccount) error
	Get(ctx context.Context, id string) (*ServiceAccount, error)
	GetByName(ctx context.Context, name string) (*ServiceAccount, error)
	List(ctx context.Context) ([]*ServiceAccount, error)
	Update(ctx context.Context, sa *ServiceAccount) error
	Delete(ctx context.Context, id string) error
}

// APIToken is a bearer token owned by a user or service account.
type APIToken struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	TokenHash  string        `json:"-"`
	OwnerID    string        `json:"owner_id"`
	OwnerType  PrincipalType `json:"owner_type"`
	Scopes     []string      `json:"scopes"`
	ExpiresAt  *time.Time    `json:"expires_at,omitempty"`
	LastUsedAt *time.Time    `json:"last_used_at,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

type APITokenRepository interface {
	Create(ctx context.Context, token *APIToken) error
	Get(ctx context.Context, id string) (*APIToken, error)
	GetByHash(ctx context.Context, hash string) (*APIToken, error)
	ListByOwner(ctx context.Context, ownerID string, ownerType PrincipalType) ([]*APIToken, error)
	UpdateLastUsed(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// Session represents an authenticated web session stored in Redis.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionStore interface {
	Create(ctx context.Context, session *Session) error
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
}
