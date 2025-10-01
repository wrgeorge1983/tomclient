package tomapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type storedToken struct {
	AccessToken string    `json:"access_token"`
	IDToken     string    `json:"id_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	ObtainedAt  time.Time `json:"obtained_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func getTokenPath(configDir string) string {
	if configDir == "" {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".tom")
	}
	return filepath.Join(configDir, "token.json")
}

func loadStoredToken(tokenPath string) (*storedToken, error) {
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not authenticated - run 'tomclient auth login' first")
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token storedToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

func (t *storedToken) IsValid() bool {
	return time.Now().Before(t.ExpiresAt.Add(-60 * time.Second))
}

func (t *storedToken) GetToken() string {
	if t.AccessToken != "" {
		return t.AccessToken
	}
	return t.IDToken
}
