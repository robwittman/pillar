package resolver

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
)

// GitHubSource resolves plugins from GitHub Releases.
//
// Source format: github.com/{owner}/{repo}
// Expected release artifacts: {repo}_{version}_{os}_{arch}.tar.gz
type GitHubSource struct {
	httpClient *http.Client
	cache      *Cache
	logger     *slog.Logger
	apiBaseURL string // overridable for testing, defaults to "https://api.github.com"
	dlBaseURL  string // overridable for testing, defaults to "https://github.com"
}

// NewGitHubSource creates a GitHub release resolver.
func NewGitHubSource(cache *Cache, logger *slog.Logger) *GitHubSource {
	return &GitHubSource{
		httpClient: http.DefaultClient,
		cache:      cache,
		logger:     logger,
		apiBaseURL: "https://api.github.com",
		dlBaseURL:  "https://github.com",
	}
}

// SetHTTPClient replaces the default HTTP client (for testing).
func (g *GitHubSource) SetHTTPClient(c *http.Client) {
	g.httpClient = c
}

// SetBaseURLs overrides the GitHub API and download base URLs (for testing).
func (g *GitHubSource) SetBaseURLs(apiBaseURL, dlBaseURL string) {
	g.apiBaseURL = apiBaseURL
	g.dlBaseURL = dlBaseURL
}

func (g *GitHubSource) Resolve(ctx context.Context, source, version string) (*Resolution, error) {
	owner, repo, err := parseGitHubSource(source)
	if err != nil {
		return nil, err
	}

	// Resolve "latest" to an actual version tag
	if version == "" || version == "latest" {
		resolved, err := g.resolveLatestVersion(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve latest version for %s: %w", source, err)
		}
		version = resolved
		g.logger.Info("resolved latest version", "source", source, "version", version)
	}

	// Check cache
	if path, ok := g.cache.Lookup(source, version); ok {
		g.logger.Info("using cached plugin", "source", source, "version", version)
		return &Resolution{BinaryPath: path, Version: version}, nil
	}

	// Download
	binary, err := g.downloadRelease(ctx, owner, repo, version)
	if err != nil {
		return nil, fmt.Errorf("failed to download plugin %s@%s: %w", source, version, err)
	}
	defer binary.Close()

	// Store in cache
	path, err := g.cache.Store(source, version, binary)
	if err != nil {
		return nil, fmt.Errorf("failed to cache plugin %s@%s: %w", source, version, err)
	}

	return &Resolution{BinaryPath: path, Version: version}, nil
}

// resolveLatestVersion queries the GitHub API for the latest release tag.
func (g *GitHubSource) resolveLatestVersion(ctx context.Context, owner, repo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", g.apiBaseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	g.addAuthHeader(req)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d for %s/%s latest release", resp.StatusCode, owner, repo)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	// Strip "v" prefix if present (v1.0.0 -> 1.0.0)
	return strings.TrimPrefix(release.TagName, "v"), nil
}

// downloadRelease downloads and extracts the plugin binary from a GitHub release.
func (g *GitHubSource) downloadRelease(ctx context.Context, owner, repo, version string) (io.ReadCloser, error) {
	assetName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", repo, version, runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("%s/%s/%s/releases/download/v%s/%s", g.dlBaseURL, owner, repo, version, assetName)

	g.logger.Info("downloading plugin", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	g.addAuthHeader(req)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download returned HTTP %d for %s", resp.StatusCode, url)
	}

	// Extract the binary from the tarball
	return g.extractBinary(resp.Body, repo)
}

// extractBinary extracts a single binary from a .tar.gz archive.
func (g *GitHubSource) extractBinary(archive io.ReadCloser, binaryName string) (io.ReadCloser, error) {
	defer archive.Close()

	gz, err := gzip.NewReader(archive)
	if err != nil {
		return nil, fmt.Errorf("failed to open gzip: %w", err)
	}

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			gz.Close()
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Match the binary by name (could be at root or in a directory)
		name := header.Name
		if i := strings.LastIndex(name, "/"); i >= 0 {
			name = name[i+1:]
		}

		if name == binaryName && header.Typeflag == tar.TypeReg {
			// Write to a temp file since we need to close the gz reader
			tmpFile, err := os.CreateTemp("", "plugin-download-*")
			if err != nil {
				gz.Close()
				return nil, err
			}
			if _, err := io.Copy(tmpFile, tr); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				gz.Close()
				return nil, err
			}
			gz.Close()
			if _, err := tmpFile.Seek(0, 0); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return nil, err
			}
			return &tempFileReadCloser{tmpFile}, nil
		}
	}

	gz.Close()
	return nil, fmt.Errorf("binary %s not found in archive", binaryName)
}

// addAuthHeader adds a GitHub token if available.
func (g *GitHubSource) addAuthHeader(req *http.Request) {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// parseGitHubSource extracts owner and repo from "github.com/owner/repo".
func parseGitHubSource(source string) (owner, repo string, err error) {
	trimmed := strings.TrimPrefix(source, "github.com/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid GitHub source %q, expected github.com/owner/repo", source)
	}
	return parts[0], parts[1], nil
}

// tempFileReadCloser wraps an os.File that deletes itself on Close.
type tempFileReadCloser struct {
	f *os.File
}

func (t *tempFileReadCloser) Read(p []byte) (int, error) {
	return t.f.Read(p)
}

func (t *tempFileReadCloser) Close() error {
	name := t.f.Name()
	t.f.Close()
	return os.Remove(name)
}
