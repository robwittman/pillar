package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// TokenPrefix is prepended to all Pillar API tokens for easy identification.
	TokenPrefix = "plt_"

	// tokenIDLen is the length of the token ID component (UUID-like).
	tokenRandomLen = 32
)

// GenerateToken creates a new API token with the format: plt_<base64(tokenID.randomBytes)>.
// Returns the raw token string (shown to user once) and its SHA-256 hash (stored in DB).
func GenerateToken(tokenID string) (rawToken string, tokenHash string, err error) {
	randomBytes := make([]byte, tokenRandomLen)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("generating random bytes: %w", err)
	}

	payload := tokenID + "." + base64.RawURLEncoding.EncodeToString(randomBytes)
	rawToken = TokenPrefix + base64.RawURLEncoding.EncodeToString([]byte(payload))
	tokenHash = HashToken(rawToken)

	return rawToken, tokenHash, nil
}

// HashToken computes the SHA-256 hash of a raw token string.
func HashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// ParseTokenID extracts the token ID from a raw token string.
// Returns empty string if the token format is invalid.
func ParseTokenID(rawToken string) string {
	if !strings.HasPrefix(rawToken, TokenPrefix) {
		return ""
	}

	encoded := strings.TrimPrefix(rawToken, TokenPrefix)
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}

	parts := strings.SplitN(string(decoded), ".", 2)
	if len(parts) != 2 {
		return ""
	}

	return parts[0]
}

// GenerateSecret creates a random secret string suitable for service account credentials.
func GenerateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateSessionID creates a random session identifier.
func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating session id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
