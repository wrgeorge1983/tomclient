package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"tomclient/internal"
)

var (
	reportInputDir  string
	reportOutputFile string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate CSV report from inventory files",
	Long: `Parse inventory files and generate a comprehensive CSV report
with device information, serial numbers, and age calculations.`,
	Example: `  tomclient report --input-dir=inventory --output=report.csv
  tomclient report -i ./data -o devices.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		err := internal.GenerateInventoryReport(reportInputDir)
		handleError(err)
		
		fmt.Printf("Inventory report generated: %s\n", reportOutputFile)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// POSIX-style flags with both long and short versions
	reportCmd.Flags().StringVarP(&reportInputDir, "input-dir", "i", "inventory", "Directory containing inventory files")
	reportCmd.Flags().StringVarP(&reportOutputFile, "output", "o", "inventory_report.csv", "Output CSV file name")
}