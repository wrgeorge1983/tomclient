package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tomclient/auth/providers"
)

type AuthMode string

const (
	AuthModeNone   AuthMode = "none"
	AuthModeAPIKey AuthMode = "api_key"
	AuthModeJWT    AuthMode = "jwt"
)

type Config struct {
	Include           string   `json:"include,omitempty"` // exclusive with all other fields
	APIURL            string   `json:"api_url,omitempty"`
	AuthMode          AuthMode `json:"auth_mode"`
	APIKey            string   `json:"api_key,omitempty"`
	APIKeyHeader      string   `json:"api_key_header,omitempty"`
	OAuthProvider     string   `json:"oauth_provider,omitempty"`
	OAuthClientID     string   `json:"oauth_client_id,omitempty"`
	OAuthClientSecret string   `json:"oauth_client_secret,omitempty"`
	OAuthDiscoveryURL string   `json:"oauth_discovery_url,omitempty"`
	OAuthRedirectPort int      `json:"oauth_redirect_port,omitempty"`
	OAuthScopes       string   `json:"oauth_scopes,omitempty"`
	OAuthUseRefresh   bool     `json:"oauth_use_refresh,omitempty"`
	CacheEnabled      bool     `json:"cache_enabled,omitempty"`
	CacheTTL          int      `json:"cache_ttl,omitempty"`
	ConfigDir         string   `json:"-"`
}

func (c *Config) GetAuthMode() string {
	return string(c.AuthMode)
}

func (c *Config) GetAPIKey() string {
	return c.APIKey
}

func (c *Config) GetAPIKeyHeader() string {
	return c.APIKeyHeader
}

func (c *Config) GetConfigDir() string {
	return c.ConfigDir
}

func (c *Config) GetCacheEnabled() bool {
	return c.CacheEnabled
}

func (c *Config) GetCacheTTL() int {
	return c.CacheTTL
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
		CacheEnabled:      true, // Enable caching by default
		CacheTTL:          300,  // 5 minutes default TTL
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
	if cfg.Include != "" {
		if cfg.Include == "config.json" {
			return nil, fmt.Errorf("config include cannot be 'config.json'")
		}
		// Validate that include file matches config-*.json pattern
		if !strings.HasPrefix(cfg.Include, "config-") || !strings.HasSuffix(cfg.Include, ".json") {
			return nil, fmt.Errorf("config include must match pattern 'config-*.json', got '%s'", cfg.Include)
		}
		includePath := filepath.Join(cfg.ConfigDir, cfg.Include)
		data, err := os.ReadFile(includePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read included config file '%s': %w", cfg.Include, err)
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse included config file '%s': %w", cfg.Include, err)
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
	if useRefresh := os.Getenv("TOM_OAUTH_USE_REFRESH"); useRefresh != "" {
		if useRefresh == "1" || useRefresh == "true" || useRefresh == "TRUE" {
			cfg.OAuthUseRefresh = true
		} else {
			cfg.OAuthUseRefresh = false
		}
	}
	if cacheEnabled := os.Getenv("TOM_CACHE_ENABLED"); cacheEnabled != "" {
		if cacheEnabled == "0" || cacheEnabled == "false" || cacheEnabled == "FALSE" {
			cfg.CacheEnabled = false
		} else {
			cfg.CacheEnabled = true
		}
	}
	if cacheTTL := os.Getenv("TOM_CACHE_TTL"); cacheTTL != "" {
		fmt.Sscanf(cacheTTL, "%d", &cfg.CacheTTL)
	}

	// Ensure offline_access is requested when refresh is enabled (not for Google)
	if cfg.OAuthUseRefresh && cfg.OAuthProvider != "google" && !strings.Contains(cfg.OAuthScopes, "offline_access") {
		if cfg.OAuthScopes == "" {
			cfg.OAuthScopes = "offline_access"
		} else {
			cfg.OAuthScopes += " offline_access"
		}
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

// ListProfiles returns a list of all config profile files in the config directory
func ListProfiles(configDir string) ([]string, error) {
	if configDir == "" {
		configDir = GetConfigDir()
	}

	entries, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Only include files matching config-*.json pattern
		if !strings.HasPrefix(name, "config-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		// Return without "config-" prefix and .json extension
		profileName := strings.TrimSuffix(strings.TrimPrefix(name, "config-"), ".json")
		profiles = append(profiles, profileName)
	}

	return profiles, nil
}

// GetCurrentProfile returns the name of the currently active profile (from include field)
func GetCurrentProfile(configDir string) (string, error) {
	if configDir == "" {
		configDir = GetConfigDir()
	}

	configPath := GetConfigPath(configDir)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg struct {
		Include string `json:"include,omitempty"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.Include == "" {
		return "", nil
	}

	// Strip "config-" prefix and ".json" extension from include value
	profileName := strings.TrimSuffix(cfg.Include, ".json")
	profileName = strings.TrimPrefix(profileName, "config-")
	return profileName, nil
}

// SetCurrentProfile updates config.json to include the specified profile
func SetCurrentProfile(configDir, profileName string) error {
	if configDir == "" {
		configDir = GetConfigDir()
	}

	// Add config- prefix if not present
	if !strings.HasPrefix(profileName, "config-") {
		profileName = "config-" + profileName
	}

	if !strings.HasSuffix(profileName, ".json") {
		profileName = profileName + ".json"
	}

	// Verify the profile file exists
	profilePath := filepath.Join(configDir, profileName)
	if _, err := os.Stat(profilePath); err != nil {
		if os.IsNotExist(err) {
			displayName := strings.TrimSuffix(strings.TrimPrefix(profileName, "config-"), ".json")
			return fmt.Errorf("profile '%s' does not exist", displayName)
		}
		return fmt.Errorf("failed to check profile file: %w", err)
	}

	// Create config.json with just the include
	includeConfig := map[string]string{
		"include": profileName,
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(includeConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := GetConfigPath(configDir)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveProfile saves a config to a specific profile file (not config.json)
func SaveProfile(cfg *Config, configDir, profileName string) error {
	if configDir == "" {
		configDir = GetConfigDir()
	}

	// Add config- prefix if not present
	if !strings.HasPrefix(profileName, "config-") {
		profileName = "config-" + profileName
	}

	if !strings.HasSuffix(profileName, ".json") {
		profileName = profileName + ".json"
	}

	if profileName == "config.json" {
		return fmt.Errorf("cannot save to 'config.json' directly, use a profile name")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Clear the Include field before saving to profile
	cfg.Include = ""

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	profilePath := filepath.Join(configDir, profileName)
	if err := os.WriteFile(profilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// LoadProfile loads a specific profile by name
func LoadProfile(configDir, profileName string) (*Config, error) {
	if configDir == "" {
		configDir = GetConfigDir()
	}

	// Add config- prefix if not present
	if !strings.HasPrefix(profileName, "config-") {
		profileName = "config-" + profileName
	}

	if !strings.HasSuffix(profileName, ".json") {
		profileName = profileName + ".json"
	}

	cfg := &Config{
		ConfigDir:         configDir,
		AuthMode:          AuthModeNone,
		OAuthProvider:     "oidc",
		OAuthRedirectPort: 8899,
		OAuthScopes:       "openid email profile",
		CacheEnabled:      true,
		CacheTTL:          300,
	}

	profilePath := filepath.Join(configDir, profileName)
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}

	return cfg, nil
}
