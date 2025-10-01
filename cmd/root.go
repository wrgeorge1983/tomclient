package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tomclient/auth"
	"tomclient/tomapi"
)

var (
	apiURL    string
	configDir string
	client    *tomapi.Client
)

var rootCmd = &cobra.Command{
	Use:   "tomclient",
	Short: "Tom Smykowski network automation client",
	Long: `A CLI client for the Tom Smykowski network automation broker service.
Authenticate with OAuth, API keys, or no auth, then execute commands on network devices or export inventory.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "auth" || cmd.Parent() != nil && cmd.Parent().Name() == "auth" {
			return
		}

		cfg, err := auth.LoadConfig(configDir)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := cfg.Validate(); err != nil {
			fmt.Printf("Configuration error: %v\n", err)
			os.Exit(1)
		}

		finalAPIURL := apiURL
		if apiURL == getDefaultAPIURL() && cfg.APIURL != "" {
			finalAPIURL = cfg.APIURL
		}

		client = tomapi.NewClient(finalAPIURL, cfg)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&apiURL, "api-url", "a", getDefaultAPIURL(), "Tom API server URL")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", auth.GetConfigDir(), "Config directory path")
}

func getDefaultAPIURL() string {
	if url := os.Getenv("TOM_API_URL"); url != "" {
		return url
	}
	return "http://localhost:8000"
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
