package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tomclient/internal"
)

var (
	parseInputDir     string
	parseOutputDir    string
	parsePattern      string
	parseDryRun       bool
	parseDetailed     bool
	parseSummaryOnly  bool
)

var parseInterfacesCmd = &cobra.Command{
	Use:   "parse-interfaces",
	Short: "Parse interface configs and generate deletion commands",
	Long: `Parse collected interface configuration files to find interfaces with specific patterns
(like 'SSN' in description) and generate deletion commands for each device.

IMPORTANT: This command only generates command files locally - it does NOT execute 
commands on remote devices. Generated files must be manually reviewed and executed.`,
	Example: `  tomclient parse-interfaces --input-dir=interfaces --pattern=SSN
  tomclient parse-interfaces -i ./interfaces -p SSN --dry-run
  tomclient parse-interfaces -i ./interfaces -p SSN -o ./deletion-commands --detailed
  tomclient parse-interfaces --summary-only --pattern=SSN`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find all interface files
		pattern := filepath.Join(parseInputDir, "*_interfaces.txt")
		files, err := filepath.Glob(pattern)
		handleError(err)

		if len(files) == 0 {
			fmt.Printf("No interface files found in %s\n", parseInputDir)
			return
		}

		// Create output directory if not dry-run and not summary-only
		if !parseDryRun && !parseSummaryOnly {
			err = os.MkdirAll(parseOutputDir, 0755)
			handleError(err)
		}

		fmt.Printf("Processing %d interface files...\n", len(files))
		if !parseSummaryOnly {
			fmt.Printf("⚠️  IMPORTANT: This generates commands locally only - NO remote execution!\n")
		}
		
		totalDevices := 0
		totalInterfaces := 0
		devicesSummary := make(map[string]int)

		// Process each file
		for _, file := range files {
			deviceInfo, err := internal.ParseInterfaceConfig(file)
			if err != nil {
				fmt.Printf("Error parsing %s: %v\n", file, err)
				continue
			}

			// Find interfaces matching pattern
			var matchingInterfaces []internal.InterfaceInfo
			for _, iface := range deviceInfo.Interfaces {
				if strings.Contains(strings.ToUpper(iface.Description), strings.ToUpper(parsePattern)) {
					matchingInterfaces = append(matchingInterfaces, iface)
				}
			}

			if len(matchingInterfaces) == 0 {
				continue
			}

			totalDevices++
			totalInterfaces += len(matchingInterfaces)
			devicesSummary[deviceInfo.Hostname] = len(matchingInterfaces)

			if parseSummaryOnly {
				fmt.Printf("  %s: %d interfaces with '%s'\n", deviceInfo.Hostname, len(matchingInterfaces), parsePattern)
				continue
			}

			// Generate deletion commands
			var commands []string
			if parseDetailed {
				commands = internal.GenerateDeleteCommandsDetailed(matchingInterfaces)
			} else {
				commands = internal.GenerateDeleteCommands(matchingInterfaces)
			}

			// Output commands
			if parseDryRun {
				fmt.Printf("\n=== %s (%d interfaces) ===\n", deviceInfo.Hostname, len(matchingInterfaces))
				for _, cmd := range commands {
					fmt.Println(cmd)
				}
			} else {
				// Write to file
				outputFile := filepath.Join(parseOutputDir, deviceInfo.Hostname+"_delete_ssn_interfaces.txt")
				content := strings.Join(commands, "\n")
				err = os.WriteFile(outputFile, []byte(content), 0644)
				if err != nil {
					fmt.Printf("Error writing commands for %s: %v\n", deviceInfo.Hostname, err)
					continue
				}
				fmt.Printf("  %s: %d interfaces -> %s\n", deviceInfo.Hostname, len(matchingInterfaces), outputFile)
			}
		}

		// Print summary
		fmt.Printf("\n=== Summary ===\n")
		fmt.Printf("Devices with '%s' interfaces: %d\n", parsePattern, totalDevices)
		fmt.Printf("Total interfaces to delete: %d\n", totalInterfaces)

		if !parseDryRun && !parseSummaryOnly {
			fmt.Printf("Deletion commands saved to: %s\n", parseOutputDir)
		}

		// Show detailed breakdown if requested
		if parseSummaryOnly && len(devicesSummary) > 0 {
			fmt.Printf("\nDetailed breakdown:\n")
			for hostname, count := range devicesSummary {
				fmt.Printf("  %s: %d interfaces\n", hostname, count)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(parseInterfacesCmd)

	// POSIX-style flags
	parseInterfacesCmd.Flags().StringVarP(&parseInputDir, "input-dir", "i", "interfaces", "Directory containing interface config files")
	parseInterfacesCmd.Flags().StringVarP(&parseOutputDir, "output-dir", "o", "deletion-commands", "Output directory for deletion command files")
	parseInterfacesCmd.Flags().StringVarP(&parsePattern, "pattern", "p", "SSN", "Pattern to search for in interface descriptions")
	parseInterfacesCmd.Flags().BoolVarP(&parseDryRun, "dry-run", "n", false, "Show what would be done without creating files")
	parseInterfacesCmd.Flags().BoolVarP(&parseDetailed, "detailed", "d", true, "Generate detailed commands with comments")
	parseInterfacesCmd.Flags().BoolVarP(&parseSummaryOnly, "summary-only", "s", false, "Only show summary, don't generate commands")
}