package resolver

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCache_StoreAndLookup(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())

	source := "github.com/example/my-plugin"
	version := "1.0.0"

	// Store a binary
	content := "#!/bin/sh\necho hello"
	path, err := cache.Store(source, version, strings.NewReader(content))
	require.NoError(t, err)
	assert.Contains(t, path, "my-plugin")

	// Verify binary was written
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	// Verify it's executable
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.Mode()&0111 != 0, "binary should be executable")

	// Lookup should find it
	found, ok := cache.Lookup(source, version)
	assert.True(t, ok)
	assert.Equal(t, path, found)
}

func TestCache_LookupMissing(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())

	path, ok := cache.Lookup("github.com/example/nonexistent", "1.0.0")
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestCache_LookupCorruptChecksum(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())

	source := "github.com/example/my-plugin"
	version := "1.0.0"

	// Store a binary
	path, err := cache.Store(source, version, strings.NewReader("binary content"))
	require.NoError(t, err)
	require.NotEmpty(t, path)

	// Corrupt the checksum file
	checksumPath := cache.checksumPath(source, version)
	require.NoError(t, os.WriteFile(checksumPath, []byte("bad-checksum"), 0644))

	// Lookup should fail due to checksum mismatch
	_, ok := cache.Lookup(source, version)
	assert.False(t, ok)
}

func TestCache_DirectoryStructure(t *testing.T) {
	cacheDir := t.TempDir()
	cache := NewCache(cacheDir, testLogger())

	source := "github.com/robwittman/pillar-plugin-keycloak"
	version := "2.1.0"

	path, err := cache.Store(source, version, strings.NewReader("binary"))
	require.NoError(t, err)

	expected := filepath.Join(cacheDir, "github.com/robwittman/pillar-plugin-keycloak/2.1.0/pillar-plugin-keycloak")
	assert.Equal(t, expected, path)
}
