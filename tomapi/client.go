package tomapi

import (
	"fmt"
	"net/http"

	"tomclient/auth"
)

type AuthConfig interface {
	GetAuthMode() string
	GetAPIKey() string
	GetConfigDir() string
}

type Client struct {
	BaseURL    string
	AuthConfig AuthConfig
	HTTPClient *http.Client
}

func NewClient(baseURL string, authConfig AuthConfig) *Client {
	return &Client{
		BaseURL:    baseURL,
		AuthConfig: authConfig,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if err := c.setAuthHeader(req); err != nil {
		return nil, err
	}

	return c.HTTPClient.Do(req)
}

func (c *Client) setAuthHeader(req *http.Request) error {
	if c.AuthConfig == nil {
		return nil
	}

	authMode := c.AuthConfig.GetAuthMode()

	switch authMode {
	case "none":
		return nil

	case "api_key":
		apiKey := c.AuthConfig.GetAPIKey()
		if apiKey == "" {
			return fmt.Errorf("auth_mode is 'api_key' but TOM_API_KEY is not set")
		}
		req.Header.Set("X-API-Key", apiKey)
		return nil

	case "jwt":
		token, err := c.loadJWTToken()
		if err != nil {
			return err
		}
		// Always use ID token for OIDC-backed auth
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil

	default:
		return fmt.Errorf("unknown auth mode: %s", authMode)
	}
}

func (c *Client) loadJWTToken() (string, error) {
	configDir := c.AuthConfig.GetConfigDir()

	t, err := auth.LoadToken(configDir)
	if err != nil {
		return "", err
	}

	// Require an ID token for OIDC
	if t.IDToken == "" {
		return "", fmt.Errorf("no id_token present; ensure 'openid' scope and re-authenticate")
	}

	if t.IsValid() { // validates ID token exp
		return t.IDToken, nil
	}

	// Attempt auto-refresh if enabled and we have a refresh token
	cfg, cfgErr := auth.LoadConfig(configDir)
	if cfgErr == nil && cfg.OAuthUseRefresh && t.RefreshToken != "" {
		newTok, rerr := auth.RefreshAccessToken(cfg, t.RefreshToken)
		if rerr == nil {
			// Persist tokens (handles rotation)
			if serr := auth.SaveToken(newTok, cfg.ConfigDir, cfg.OAuthProvider); serr == nil {
				// Reload and return ID token
				if latest, lerr := auth.LoadToken(configDir); lerr == nil && latest.IDToken != "" && latest.IsValid() {
					return latest.IDToken, nil
				}
			}
		}
	}

	return "", fmt.Errorf("token expired - run 'tomclient auth login' to re-authenticate")
}
