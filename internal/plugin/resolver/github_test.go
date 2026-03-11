package resolver

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGitHubSource(t *testing.T) {
	tests := []struct {
		source string
		owner  string
		repo   string
		err    bool
	}{
		{"github.com/robwittman/pillar-plugin-keycloak", "robwittman", "pillar-plugin-keycloak", false},
		{"github.com/org/repo", "org", "repo", false},
		{"github.com/invalid", "", "", true},
		{"github.com//repo", "", "", true},
		{"github.com/owner/", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			owner, repo, err := parseGitHubSource(tt.source)
			if tt.err {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.owner, owner)
				assert.Equal(t, tt.repo, repo)
			}
		})
	}
}

func makeTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: name,
		Mode: 0755,
		Size: int64(len(content)),
	}
	require.NoError(t, tw.WriteHeader(hdr))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())
	return buf.Bytes()
}

func TestGitHubSource_ExtractBinary(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho plugin")
	tarball := makeTarGz(t, "my-plugin", binaryContent)

	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())
	src := NewGitHubSource(cache, testLogger())

	reader, err := src.extractBinary(
		&readCloserWrapper{bytes.NewReader(tarball)},
		"my-plugin",
	)
	require.NoError(t, err)
	defer reader.Close()

	extracted, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, binaryContent, extracted)
}

func TestGitHubSource_ExtractBinaryNotFound(t *testing.T) {
	tarball := makeTarGz(t, "wrong-name", []byte("content"))

	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())
	src := NewGitHubSource(cache, testLogger())

	_, err := src.extractBinary(
		&readCloserWrapper{bytes.NewReader(tarball)},
		"expected-name",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in archive")
}

func TestGitHubSource_ResolveCached(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())
	src := NewGitHubSource(cache, testLogger())

	// Pre-cache a binary
	path, err := cache.Store("github.com/owner/my-plugin", "1.2.0", bytes.NewReader([]byte("cached")))
	require.NoError(t, err)

	res, err := src.Resolve(t.Context(), "github.com/owner/my-plugin", "1.2.0")
	require.NoError(t, err)
	assert.Equal(t, path, res.BinaryPath)
	assert.Equal(t, "1.2.0", res.Version)
}

func TestGitHubSource_FullDownloadResolve(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho plugin")
	repo := "my-plugin"
	version := "1.0.0"
	assetName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", repo, version, runtime.GOOS, runtime.GOARCH)
	tarball := makeTarGz(t, repo, binaryContent)

	mux := http.NewServeMux()
	mux.HandleFunc(
		fmt.Sprintf("/owner/%s/releases/download/v%s/%s", repo, version, assetName),
		func(w http.ResponseWriter, r *http.Request) {
			w.Write(tarball)
		},
	)

	server := httptest.NewServer(mux)
	defer server.Close()

	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())
	src := NewGitHubSource(cache, testLogger())
	src.SetHTTPClient(server.Client())
	src.SetBaseURLs(server.URL, server.URL)

	res, err := src.Resolve(t.Context(), "github.com/owner/my-plugin", version)
	require.NoError(t, err)
	assert.Equal(t, version, res.Version)
	assert.NotEmpty(t, res.BinaryPath)

	// Verify the binary content was cached correctly
	data, err := os.ReadFile(res.BinaryPath)
	require.NoError(t, err)
	assert.Equal(t, binaryContent, data)

	// Second resolve should use cache (no HTTP needed)
	server.Close()
	src2 := NewGitHubSource(cache, testLogger())
	res2, err := src2.Resolve(t.Context(), "github.com/owner/my-plugin", version)
	require.NoError(t, err)
	assert.Equal(t, res.BinaryPath, res2.BinaryPath)
}

func TestGitHubSource_ResolveLatestVersion(t *testing.T) {
	binaryContent := []byte("#!/bin/sh\necho v2")
	repo := "my-plugin"
	version := "2.0.0"
	assetName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", repo, version, runtime.GOOS, runtime.GOARCH)
	tarball := makeTarGz(t, repo, binaryContent)

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/my-plugin/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"tag_name": "v2.0.0"})
	})
	mux.HandleFunc(
		fmt.Sprintf("/owner/%s/releases/download/v%s/%s", repo, version, assetName),
		func(w http.ResponseWriter, r *http.Request) {
			w.Write(tarball)
		},
	)

	server := httptest.NewServer(mux)
	defer server.Close()

	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())
	src := NewGitHubSource(cache, testLogger())
	src.SetHTTPClient(server.Client())
	src.SetBaseURLs(server.URL, server.URL)

	res, err := src.Resolve(t.Context(), "github.com/owner/my-plugin", "latest")
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", res.Version)
	assert.NotEmpty(t, res.BinaryPath)
}

func TestGitHubSource_DownloadError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())
	src := NewGitHubSource(cache, testLogger())
	src.SetHTTPClient(server.Client())
	src.SetBaseURLs(server.URL, server.URL)

	_, err := src.Resolve(t.Context(), "github.com/owner/nonexistent", "1.0.0")
	assert.Error(t, err)
}

// readCloserWrapper wraps a bytes.Reader as an io.ReadCloser.
type readCloserWrapper struct {
	r *bytes.Reader
}

func (w *readCloserWrapper) Read(p []byte) (int, error) { return w.r.Read(p) }
func (w *readCloserWrapper) Close() error               { return nil }
