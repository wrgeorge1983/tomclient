package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tomclient/tomapi"
)

var (
	apiURL string
	client *tomapi.Client
)

var rootCmd = &cobra.Command{
	Use:   "tomclient",
	Short: "Tom Smykowski network automation client",
	Long: `A CLI client for the Tom Smykowski network automation broker service.
Supports device command execution, inventory management, and bulk operations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize client with the API URL
		client = tomapi.NewClient(apiURL)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&apiURL, "api-url", "a", getDefaultAPIURL(), "Tom API server URL")
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