package resolver

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Cache manages locally cached plugin binaries.
type Cache struct {
	baseDir string
	logger  *slog.Logger
}

// NewCache creates a cache rooted at the given directory.
func NewCache(baseDir string, logger *slog.Logger) *Cache {
	return &Cache{
		baseDir: baseDir,
		logger:  logger,
	}
}

// Lookup checks if a plugin binary exists in the cache for the given source and version.
// Returns the binary path if found, or empty string if not cached.
func (c *Cache) Lookup(source, version string) (string, bool) {
	binPath := c.binaryPath(source, version)
	if _, err := os.Stat(binPath); err != nil {
		return "", false
	}

	// Verify checksum if available
	checksumPath := c.checksumPath(source, version)
	if expected, err := os.ReadFile(checksumPath); err == nil {
		actual, err := fileChecksum(binPath)
		if err != nil || actual != strings.TrimSpace(string(expected)) {
			c.logger.Warn("cached plugin checksum mismatch, will re-download",
				"source", source, "version", version)
			return "", false
		}
	}

	return binPath, true
}

// Store writes a plugin binary to the cache and records its checksum.
func (c *Cache) Store(source, version string, binary io.Reader) (string, error) {
	dir := c.versionDir(source, version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache dir: %w", err)
	}

	binPath := c.binaryPath(source, version)

	// Write to temp file then rename for atomicity
	tmpFile, err := os.CreateTemp(dir, "plugin-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)

	if _, err := io.Copy(writer, binary); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write plugin binary: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return "", fmt.Errorf("failed to set executable permissions: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), binPath); err != nil {
		return "", fmt.Errorf("failed to move plugin binary to cache: %w", err)
	}

	// Write checksum
	checksum := hex.EncodeToString(hasher.Sum(nil))
	checksumPath := c.checksumPath(source, version)
	if err := os.WriteFile(checksumPath, []byte(checksum), 0644); err != nil {
		c.logger.Warn("failed to write checksum file", "error", err)
	}

	c.logger.Info("plugin cached", "source", source, "version", version, "path", binPath)
	return binPath, nil
}

// versionDir returns the cache directory for a specific source+version.
func (c *Cache) versionDir(source, version string) string {
	return filepath.Join(c.baseDir, source, version)
}

// binaryPath returns the expected binary path for a source+version.
func (c *Cache) binaryPath(source, version string) string {
	// Binary name is the last component of the source path
	parts := strings.Split(source, "/")
	binaryName := parts[len(parts)-1]
	return filepath.Join(c.versionDir(source, version), binaryName)
}

// checksumPath returns the checksum file path for a source+version.
func (c *Cache) checksumPath(source, version string) string {
	return filepath.Join(c.versionDir(source, version), "SHA256SUM")
}

// fileChecksum computes the SHA256 checksum of a file.
func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
