package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHubProvider authenticates users via GitHub OAuth2.
// GitHub is not OIDC-compliant, so we use the GitHub user API directly.
type GitHubProvider struct {
	name   string
	oauth  oauth2.Config
	client *http.Client
}

// GitHubProviderConfig holds the configuration for a GitHub provider.
type GitHubProviderConfig struct {
	Name         string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// GitHubUserAPIURL is the endpoint used to fetch GitHub user info.
// Exported so tests can override it.
var GitHubUserAPIURL = "https://api.github.com/user"

func NewGitHubProvider(cfg GitHubProviderConfig) *GitHubProvider {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read:user", "user:email"}
	}

	return &GitHubProvider{
		name: cfg.Name,
		oauth: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     github.Endpoint,
			Scopes:       scopes,
		},
		client: http.DefaultClient,
	}
}

func (p *GitHubProvider) Type() ProviderType { return ProviderGitHub }
func (p *GitHubProvider) Name() string       { return p.name }

func (p *GitHubProvider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state)
}

func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*ExternalIdentity, error) {
	token, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GitHubUserAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching github user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github user API returned %d: %s", resp.StatusCode, body)
	}

	var ghUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, fmt.Errorf("decoding github user: %w", err)
	}

	displayName := ghUser.Name
	if displayName == "" {
		displayName = ghUser.Login
	}

	return &ExternalIdentity{
		Provider:    p.name,
		SubjectID:   strconv.Itoa(ghUser.ID),
		Email:       ghUser.Email,
		DisplayName: displayName,
	}, nil
}

func (p *GitHubProvider) ValidateCredentials(_ context.Context, _, _ string) (*ExternalIdentity, error) {
	return nil, fmt.Errorf("GitHub provider does not support password authentication")
}
