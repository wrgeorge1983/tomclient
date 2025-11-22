package auth

import (
	"fmt"
	"net/http"
)

// CLIAuthProvider implements AuthProvider for CLI using config and envvars
type CLIAuthProvider struct {
	config *Config
}

func NewCLIAuthProvider(cfg *Config) *CLIAuthProvider {
	return &CLIAuthProvider{config: cfg}
}

func (p *CLIAuthProvider) AddAuth(req *http.Request) error {
	switch p.config.AuthMode {
	case AuthModeNone:
		return nil

	case AuthModeAPIKey:
		key := p.config.APIKey
		if key == "" {
			return fmt.Errorf("auth_mode is 'api_key' but API key is not set")
		}
		header := "X-API-Key"
		if p.config.APIKeyHeader != "" {
			header = p.config.APIKeyHeader
		}
		req.Header.Set(header, key)
		return nil

	case AuthModeJWT:
		token, err := p.loadJWTToken()
		if err != nil {
			return fmt.Errorf("failed to get JWT token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nil

	default:
		return fmt.Errorf("invalid auth_mode '%s'", p.config.AuthMode)
	}
}

func (p *CLIAuthProvider) loadJWTToken() (string, error) {
	t, err := LoadToken(p.config.ConfigDir)
	if err != nil {
		return "", fmt.Errorf("failed to load token: %w", err)
	}

	// Require an ID token for OIDC
	if t.IDToken == "" {
		return "", fmt.Errorf("no id_token present; ensure 'openid' scope and re-authenticate")
	}

	if t.IsValid() {
		return t.IDToken, nil
	}

	if p.config.OAuthUseRefresh && t.RefreshToken != "" {
		refreshResponse, err := RefreshAccessToken(p.config, t.RefreshToken)
		if err != nil {
			return "", fmt.Errorf("failed to refresh access token: %w", err)
		}
		err = SaveToken(refreshResponse, p.config.ConfigDir, p.config.OAuthProvider)
		if err != nil {
			return "", fmt.Errorf("failed to save refreshed token: %w", err)
		}
		return refreshResponse.IDToken, nil
	}
	return "", fmt.Errorf("token expired and no refresh token available; please re-authenticate with 'tom auth login'")
}
