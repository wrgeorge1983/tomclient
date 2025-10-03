package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"tomclient/auth/providers"
)

type AuthMode string

const (
	AuthModeNone   AuthMode = "none"
	AuthModeAPIKey AuthMode = "api_key"
	AuthModeJWT    AuthMode = "jwt"
)

type Config struct {
	APIURL            string   `json:"api_url,omitempty"`
	AuthMode          AuthMode `json:"auth_mode"`
	APIKey            string   `json:"api_key,omitempty"`
	OAuthProvider     string   `json:"oauth_provider,omitempty"`
	OAuthClientID     string   `json:"oauth_client_id,omitempty"`
	OAuthClientSecret string   `json:"oauth_client_secret,omitempty"`
	OAuthDiscoveryURL string   `json:"oauth_discovery_url,omitempty"`
	OAuthRedirectPort int      `json:"oauth_redirect_port,omitempty"`
	OAuthScopes       string   `json:"oauth_scopes,omitempty"`
	ConfigDir         string   `json:"-"`
}

func (c *Config) GetAuthMode() string {
	return string(c.AuthMode)
}

func (c *Config) GetAPIKey() string {
	return c.APIKey
}

func (c *Config) GetConfigDir() string {
	return c.ConfigDir
}

func GetConfigDir() string {
	if dir := os.Getenv("TOM_CONFIG_DIR"); dir != "" {
		return dir
	}

	// If running under sudo, use the original user's home directory
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		if sudoHome := os.Getenv("SUDO_HOME"); sudoHome != "" {
			return filepath.Join(sudoHome, ".tom")
		}
		// Fallback: construct home directory from SUDO_USER
		return filepath.Join("/home", sudoUser, ".tom")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".tom"
	}
	return filepath.Join(homeDir, ".tom")
}

func GetConfigPath(configDir string) string {
	if configDir == "" {
		configDir = GetConfigDir()
	}
	return filepath.Join(configDir, "config.json")
}

func LoadConfig(configDir string) (*Config, error) {
	cfg := &Config{
		ConfigDir:         configDir,
		AuthMode:          AuthModeNone,
		OAuthProvider:     "oidc",
		OAuthRedirectPort: 8899,
		OAuthScopes:       "openid email profile",
	}

	if cfg.ConfigDir == "" {
		cfg.ConfigDir = GetConfigDir()
	}

	configPath := GetConfigPath(cfg.ConfigDir)
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	if apiURL := os.Getenv("TOM_API_URL"); apiURL != "" {
		cfg.APIURL = apiURL
	}
	if authMode := os.Getenv("TOM_AUTH_MODE"); authMode != "" {
		cfg.AuthMode = AuthMode(authMode)
	}
	if apiKey := os.Getenv("TOM_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if provider := os.Getenv("TOM_OAUTH_PROVIDER"); provider != "" {
		cfg.OAuthProvider = provider
	}
	if clientID := os.Getenv("TOM_OAUTH_CLIENT_ID"); clientID != "" {
		cfg.OAuthClientID = clientID
	}
	if clientSecret := os.Getenv("TOM_OAUTH_CLIENT_SECRET"); clientSecret != "" {
		cfg.OAuthClientSecret = clientSecret
	}
	if discoveryURL := os.Getenv("TOM_OAUTH_DISCOVERY_URL"); discoveryURL != "" {
		cfg.OAuthDiscoveryURL = discoveryURL
	}
	if port := os.Getenv("TOM_OAUTH_REDIRECT_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.OAuthRedirectPort)
	}
	if scopes := os.Getenv("TOM_OAUTH_SCOPES"); scopes != "" {
		cfg.OAuthScopes = scopes
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	switch c.AuthMode {
	case AuthModeNone:
		if c.APIKey != "" {
			fmt.Println("Warning: TOM_API_KEY is set but auth_mode is 'none' - API key will not be used")
		}
		if c.OAuthClientID != "" || c.OAuthDiscoveryURL != "" {
			fmt.Println("Warning: OAuth config is set but auth_mode is 'none' - OAuth will not be used")
		}
		return nil

	case AuthModeAPIKey:
		if c.APIKey == "" {
			return fmt.Errorf("auth_mode is 'api_key' but TOM_API_KEY is not set")
		}
		if c.OAuthClientID != "" || c.OAuthDiscoveryURL != "" {
			fmt.Println("Warning: OAuth config is set but auth_mode is 'api_key' - OAuth will not be used")
		}
		return nil

	case AuthModeJWT:
		if c.OAuthClientID == "" {
			return fmt.Errorf("auth_mode is 'jwt' but TOM_OAUTH_CLIENT_ID is not set")
		}
		if c.OAuthDiscoveryURL == "" {
			return fmt.Errorf("auth_mode is 'jwt' but TOM_OAUTH_DISCOVERY_URL is not set")
		}
		if c.APIKey != "" {
			fmt.Println("Warning: TOM_API_KEY is set but auth_mode is 'jwt' - API key will not be used")
		}

		provider, err := providers.GetProvider(c.OAuthProvider)
		if err != nil {
			return err
		}

		if provider.RequiresClientSecret() && c.OAuthClientSecret == "" {
			return fmt.Errorf("OAuth provider '%s' requires client_secret but TOM_OAUTH_CLIENT_SECRET is not set", c.OAuthProvider)
		}

		if !provider.RequiresClientSecret() && c.OAuthClientSecret != "" {
			fmt.Printf("Warning: TOM_OAUTH_CLIENT_SECRET is set but provider '%s' does not require it - client secret will be ignored\n", c.OAuthProvider)
		}

		return nil

	default:
		return fmt.Errorf("invalid auth_mode '%s' - must be one of: none, api_key, jwt", c.AuthMode)
	}
}

func (c *Config) Save() error {
	if err := os.MkdirAll(c.ConfigDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := GetConfigPath(c.ConfigDir)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
