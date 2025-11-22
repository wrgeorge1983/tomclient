package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"tomclient/auth"
)

var (
	configFromProfile string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration profiles",
	Long: `Manage configuration profiles for different environments.

Profiles allow you to maintain multiple configurations (prod, staging, local, etc.)
and switch between them easily. The main config.json file includes a reference to
the active profile.

Example structure:
  ~/.tom/
    config.json        # {"include": "config-prod.json"}
    config-prod.json   # Production configuration
    config-local.json  # Local development configuration
    config-staging.json # Staging configuration
`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available configuration profiles",
	Long:  `List all available configuration profiles in the config directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := auth.ListProfiles(configDir)
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		currentProfile, err := auth.GetCurrentProfile(configDir)
		if err != nil {
			return fmt.Errorf("failed to get current profile: %w", err)
		}

		if len(profiles) == 0 {
			fmt.Println("No configuration profiles found.")
			fmt.Printf("\nCreate a profile with: tomclient config create <name>\n")
			return nil
		}

		fmt.Println("Available configuration profiles:")
		for _, profile := range profiles {
			if profile == currentProfile {
				fmt.Printf("  * %s (active)\n", profile)
			} else {
				fmt.Printf("    %s\n", profile)
			}
		}

		if currentProfile == "" {
			fmt.Println("\nNo profile currently active. Set one with: tomclient config use <name>")
		}

		return nil
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use <profile-name>",
	Short: "Switch to a different configuration profile",
	Long:  `Switch to a different configuration profile by updating config.json to include it.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		if err := auth.SetCurrentProfile(configDir, profileName); err != nil {
			return err
		}

		fmt.Printf("Switched to profile: %s\n", profileName)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration",
	Long:  `Display the current active configuration (with environment variables applied).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := auth.LoadConfig(configDir)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get the current profile name
		currentProfile, err := auth.GetCurrentProfile(configDir)
		if err != nil {
			return fmt.Errorf("failed to get current profile: %w", err)
		}

		if currentProfile != "" {
			fmt.Printf("Active profile: %s\n\n", currentProfile)
		} else {
			fmt.Println("No profile active (using config.json directly)\n")
		}

		// Marshal and display the config
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format config: %w", err)
		}

		fmt.Println(string(data))
		return nil
	},
}

var configCreateCmd = &cobra.Command{
	Use:   "create <profile-name>",
	Short: "Create a new configuration profile",
	Long: `Create a new configuration profile.

By default, creates an empty profile with default values.
Use --from to copy settings from an existing profile.

Examples:
  tomclient config create local
  tomclient config create staging --from prod
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		if configDir == "" {
			configDir = auth.GetConfigDir()
		}

		var cfg *auth.Config
		var err error

		if configFromProfile != "" {
			// Load from existing profile
			cfg, err = auth.LoadProfile(configDir, configFromProfile)
			if err != nil {
				return fmt.Errorf("failed to load source profile '%s': %w", configFromProfile, err)
			}
			fmt.Printf("Creating profile '%s' from '%s'\n", profileName, configFromProfile)
		} else {
			// Create new with defaults
			cfg = &auth.Config{
				ConfigDir:         configDir,
				APIURL:            "http://localhost:8020",
				AuthMode:          auth.AuthModeNone,
				OAuthProvider:     "oidc",
				OAuthRedirectPort: 8899,
				OAuthScopes:       "openid email profile",
				CacheEnabled:      true,
				CacheTTL:          300,
			}
			fmt.Printf("Creating profile '%s' with default settings\n", profileName)
		}

		if err := auth.SaveProfile(cfg, configDir, profileName); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}

		fmt.Printf("Profile '%s' created successfully\n", profileName)
		fmt.Printf("\nActivate it with: tomclient config use %s\n", profileName)
		fmt.Printf("Edit at: %s/config-%s.json\n", configDir, profileName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configCreateCmd)

	configCreateCmd.Flags().StringVar(&configFromProfile, "from", "", "Copy settings from an existing profile")
}
