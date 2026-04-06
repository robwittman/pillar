package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider authenticates users via any OIDC-compliant identity provider
// (Okta, Keycloak, Auth0, etc.).
type OIDCProvider struct {
	name     string
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    oauth2.Config
}

// OIDCProviderConfig holds the configuration for an OIDC provider.
type OIDCProviderConfig struct {
	Name         string
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func NewOIDCProvider(ctx context.Context, cfg OIDCProviderConfig) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discovering OIDC provider %s: %w", cfg.Name, err)
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	oauthCfg := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	return &OIDCProvider{
		name:     cfg.Name,
		provider: provider,
		verifier: verifier,
		oauth:    oauthCfg,
	}, nil
}

func (p *OIDCProvider) Type() ProviderType { return ProviderOIDC }
func (p *OIDCProvider) Name() string       { return p.name }

func (p *OIDCProvider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state)
}

func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*ExternalIdentity, error) {
	token, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verifying id_token: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parsing claims: %w", err)
	}

	return &ExternalIdentity{
		Provider:    p.name,
		SubjectID:   claims.Sub,
		Email:       claims.Email,
		DisplayName: claims.Name,
	}, nil
}

func (p *OIDCProvider) ValidateCredentials(_ context.Context, _, _ string) (*ExternalIdentity, error) {
	return nil, fmt.Errorf("OIDC provider does not support password authentication")
}
