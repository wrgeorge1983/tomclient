package cmd

import (
	"fmt"
	"time"

	"tomclient/auth"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Manage authentication for the Tom API (OAuth, API keys, etc.)`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with OAuth",
	Long: `Authenticate with the Tom API using OAuth/PKCE flow.
Opens a browser window for authentication with your identity provider.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cfg, err := auth.LoadConfig(configDir)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		if cfg.AuthMode != auth.AuthModeJWT {
			return fmt.Errorf("auth mode is '%s' but 'auth login' requires auth_mode='jwt'\nSet TOM_AUTH_MODE=jwt or update your config file", cfg.AuthMode)
		}

		return auth.Authenticate(cfg)
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Display current authentication configuration and token status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cfg, err := auth.LoadConfig(configDir)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Printf("Auth Mode: %s\n", cfg.AuthMode)
		fmt.Printf("Config Dir: %s\n", cfg.ConfigDir)

		switch cfg.AuthMode {
		case auth.AuthModeNone:
			fmt.Println("Status: No authentication configured")

		case auth.AuthModeAPIKey:
			if cfg.APIKey != "" {
				fmt.Printf("API Key: %s***\n", cfg.APIKey[:8])
				fmt.Println("Status: API key configured")
			} else {
				fmt.Println("Status: ❌ API key not set (TOM_API_KEY required)")
			}

		case auth.AuthModeJWT:
			if cfg.OAuthClientID == "" || cfg.OAuthDiscoveryURL == "" {
				fmt.Println("Status: ❌ OAuth configuration incomplete")
				fmt.Println("Required: TOM_OAUTH_CLIENT_ID and TOM_OAUTH_DISCOVERY_URL")
				return nil
			}

			fmt.Printf("OAuth Client ID: %s\n", cfg.OAuthClientID)
			fmt.Printf("OAuth Discovery URL: %s\n", cfg.OAuthDiscoveryURL)

			token, err := auth.LoadToken(cfg.ConfigDir)
			if err != nil {
				fmt.Println("Status: ❌ Not authenticated - run 'tomclient auth login'")
				return nil
			}

			if token.IsValid() {
				expiresIn := time.Until(token.ExpiresAt).Round(time.Second)
				fmt.Printf("Status: ✅ Authenticated (expires in %v)\n", expiresIn)
				fmt.Printf("Token Type: %s\n", token.TokenType)
			} else {
				fmt.Println("Status: ❌ Token expired - run 'tomclient auth login'")
			}

			// Refresh token status
			if token.RefreshToken == "" {
				fmt.Println("Refresh Token: absent")
			} else if token.RefreshExpiresAt.IsZero() {
				fmt.Println("Refresh Token: present (expiry: unknown)")
			} else if time.Now().Before(token.RefreshExpiresAt) {
				fmt.Printf("Refresh Token: present (expires in %v)\n", time.Until(token.RefreshExpiresAt).Round(time.Second))
			} else {
				fmt.Println("Refresh Token: present (expired)")
			}
		}

		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored authentication token",
	Long:  `Remove the stored OAuth token. You will need to run 'auth login' again to authenticate.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cfg, err := auth.LoadConfig(configDir)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := auth.DeleteToken(cfg.ConfigDir); err != nil {
			return err
		}

		fmt.Println("✅ Logged out successfully")
		return nil
	},
}

var authRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record JWT token for testing",
	Long:  `Send a request to the Tom API to record JWT token for testing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cfg, err := auth.LoadConfig(configDir)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		client := createClient(getAPIURL(cfg), cfg)
		if err := client.RecordJWT(); err != nil {
			return fmt.Errorf("failed to record JWT: %w", err)
		}

		fmt.Println("✅ JWT recorded successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authRecordCmd)
}
