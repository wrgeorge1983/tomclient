package cmd

import (
	"fmt"
	
	"github.com/spf13/cobra"
	"tomclient/internal"
)

var (
	bulkConcurrency int
	bulkOutputDir   string
)

var bulkInventoryCmd = &cobra.Command{
	Use:   "bulk-inventory <devices-file>",
	Short: "Run inventory command on multiple devices",
	Long: `Execute 'show inventory | i ASR' command on all devices listed in the JSON file.
Supports concurrent execution with configurable worker count.`,
	Example: `  tomclient bulk-inventory devices.json --concurrency=20
  tomclient bulk-inventory devices.json -c 10 --output-dir=./inventory-data
  tomclient bulk-inventory devices.json -c 5 -o /tmp/inventory`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		devicesFile := args[0]

		err := internal.BulkInventory(client, devicesFile, bulkConcurrency)
		handleError(err)
	},
}

func init() {
	rootCmd.AddCommand(bulkInventoryCmd)

	// POSIX-style flags with both long and short versions
	bulkInventoryCmd.Flags().IntVarP(&bulkConcurrency, "concurrency", "c", 20, "Number of concurrent workers")
	bulkInventoryCmd.Flags().StringVarP(&bulkOutputDir, "output-dir", "o", "inventory", "Output directory for inventory files")

	// Validation
	bulkInventoryCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if bulkConcurrency < 1 {
			return fmt.Errorf("concurrency must be at least 1, got %d", bulkConcurrency)
		}
		return nil
	}
}