package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type StoredToken struct {
	AccessToken string    `json:"access_token"`
	IDToken     string    `json:"id_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	ObtainedAt  time.Time `json:"obtained_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Provider    string    `json:"provider,omitempty"`
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

	stored := StoredToken{
		AccessToken: token.AccessToken,
		IDToken:     token.IDToken,
		TokenType:   token.TokenType,
		ExpiresIn:   token.ExpiresIn,
		ObtainedAt:  time.Now(),
		ExpiresAt:   time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		Provider:    provider,
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

func (t *StoredToken) IsValid() bool {
	return time.Now().Before(t.ExpiresAt.Add(-60 * time.Second))
}

func (t *StoredToken) GetToken() string {
	if t.Provider == "google" || t.Provider == "microsoft" {
		if t.IDToken != "" {
			return t.IDToken
		}
	}

	if t.AccessToken != "" {
		return t.AccessToken
	}
	return t.IDToken
}

func DeleteToken(configDir string) error {
	tokenPath := GetTokenPath(configDir)
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}
