package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	exportFilter string
	exportFormat string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export inventory from Tom API",
	Long: `Export device inventory in various formats.
Optionally filter the results using predefined filters.`,
	Example: `  tomclient export --filter=routers --format=json
  tomclient export -f switches -o pretty`,
	Run: func(cmd *cobra.Command, args []string) {
		inventory, err := client.ExportInventory(exportFilter)
		handleError(err)

		switch exportFormat {
		case "json":
			data, err := json.Marshal(inventory)
			handleError(err)
			fmt.Println(string(data))
		case "pretty":
			prettyJSON, err := json.MarshalIndent(inventory, "", "  ")
			handleError(err)
			fmt.Println(string(prettyJSON))
		default:
			handleError(fmt.Errorf("unsupported format: %s", exportFormat))
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// POSIX-style flags with both long and short versions
	exportCmd.Flags().StringVarP(&exportFilter, "filter", "f", "", "Filter name to apply (optional)")
	exportCmd.Flags().StringVarP(&exportFormat, "output", "o", "pretty", "Output format: json, pretty")
}