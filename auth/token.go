package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	IDToken               string `json:"id_token"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in,omitempty"`
	RefreshExpiresIn      int    `json:"refresh_expires_in,omitempty"`
}

type StoredToken struct {
	AccessToken      string    `json:"access_token"`
	IDToken          string    `json:"id_token"`
	TokenType        string    `json:"token_type"`
	ExpiresIn        int       `json:"expires_in"`
	ObtainedAt       time.Time `json:"obtained_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshToken     string    `json:"refresh_token,omitempty"`
	RefreshExpiresIn int       `json:"refresh_expires_in,omitempty"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at,omitempty"`
	Provider         string    `json:"provider,omitempty"`
}

func GetTokenPath(configDir string) string {
	if configDir == "" {
		configDir = GetConfigDir()
	}
	return filepath.Join(configDir, "token.json")
}

func SaveToken(token *TokenResponse, configDir string, provider string) error {
	if configDir == "" {
		configDir = GetConfigDir()
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Preserve existing refresh_token if server does not return a new one
	var existingRefresh string
	var existingRefreshExpiresAt time.Time
	if existing, err := LoadToken(configDir); err == nil && existing != nil {
		existingRefresh = existing.RefreshToken
		existingRefreshExpiresAt = existing.RefreshExpiresAt
	}

	refresh := token.RefreshToken
	if refresh == "" {
		refresh = existingRefresh
	}

	// Determine refresh expiry
	refreshExpiresIn := token.RefreshTokenExpiresIn
	if refreshExpiresIn == 0 {
		refreshExpiresIn = token.RefreshExpiresIn
	}
	var refreshExpiresAt time.Time
	if refreshExpiresIn > 0 {
		refreshExpiresAt = time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)
	} else if refresh != "" && refresh == existingRefresh {
		refreshExpiresAt = existingRefreshExpiresAt
	}

	stored := StoredToken{
		AccessToken:      token.AccessToken,
		IDToken:          token.IDToken,
		TokenType:        token.TokenType,
		ExpiresIn:        token.ExpiresIn,
		ObtainedAt:       time.Now(),
		ExpiresAt:        time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		Provider:         provider,
		RefreshToken:     refresh,
		RefreshExpiresIn: refreshExpiresIn,
		RefreshExpiresAt: refreshExpiresAt,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	tokenPath := GetTokenPath(configDir)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func LoadToken(configDir string) (*StoredToken, error) {
	tokenPath := GetTokenPath(configDir)

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated - run 'tomclient auth login' first")
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token StoredToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

// IsValid returns whether the currently used token is valid.
// For OIDC we use the ID token; validate by its exp when present.
func (t *StoredToken) IsValid() bool {
	if t.IDToken != "" {
		if exp, ok := parseJWTExp(t.IDToken); ok {
			return time.Now().Before(exp.Add(-60 * time.Second))
		}
		// If we cannot parse, be conservative and treat as expired
		return false
	}
	// Fallback: access token expiry
	return time.Now().Before(t.ExpiresAt.Add(-60 * time.Second))
}

// GetToken returns the ID token; we no longer fall back to access tokens.
func (t *StoredToken) GetToken() string {
	return t.IDToken
}

// parseJWTExp extracts the exp claim from a JWT without verification.
func parseJWTExp(jwt string) (time.Time, bool) {
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, false
	}
	if claims.Exp == 0 {
		return time.Time{}, false
	}
	return time.Unix(claims.Exp, 0), true
}

func DeleteToken(configDir string) error {
	tokenPath := GetTokenPath(configDir)
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}
