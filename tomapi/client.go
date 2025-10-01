package tomapi

import (
	"fmt"
	"net/http"
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
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil

	default:
		return fmt.Errorf("unknown auth mode: %s", authMode)
	}
}

func (c *Client) loadJWTToken() (string, error) {
	configDir := c.AuthConfig.GetConfigDir()

	tokenPath := getTokenPath(configDir)
	token, err := loadStoredToken(tokenPath)
	if err != nil {
		return "", err
	}

	if !token.IsValid() {
		return "", fmt.Errorf("token expired - run 'tomclient auth login' to re-authenticate")
	}

	return token.GetToken(), nil
}
