package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	interfacesDevicesFile   string
	interfacesOutputDir     string
	interfacesConcurrency   int
	interfacesFilterRouters bool
	interfacesCommand       string
)

var collectInterfacesCmd = &cobra.Command{
	Use:   "collect-interfaces <devices-file>",
	Short: "Collect interface configurations from network devices",
	Long: `Collect interface configurations from routers for later processing.
Supports filtering to routers only and concurrent execution.`,
	Example: `  tomclient collect-interfaces devices.json --routers-only
  tomclient collect-interfaces devices.json -r -c 10 --output-dir=./interfaces
  tomclient collect-interfaces devices.json --command="show running-config interface"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interfacesDevicesFile = args[0]

		// Read devices file
		data, err := os.ReadFile(interfacesDevicesFile)
		handleError(err)

		var devices map[string]interface{}
		err = json.Unmarshal(data, &devices)
		handleError(err)

		// Filter devices if routers-only is set
		var targetDevices []string
		for hostname := range devices {
			if interfacesFilterRouters {
				if isRouter(hostname) {
					targetDevices = append(targetDevices, hostname)
				}
			} else {
				targetDevices = append(targetDevices, hostname)
			}
		}

		if len(targetDevices) == 0 {
			fmt.Println("No devices found matching criteria")
			return
		}

		// Create output directory
		err = os.MkdirAll(interfacesOutputDir, 0755)
		handleError(err)

		fmt.Printf("Collecting interface configs from %d devices with %d concurrent workers...\n", 
			len(targetDevices), interfacesConcurrency)

		// Execute concurrent collection
		sem := make(chan struct{}, interfacesConcurrency)
		var wg sync.WaitGroup

		for _, hostname := range targetDevices {
			wg.Add(1)
			go func(h string) {
				sem <- struct{}{}
				collectDeviceInterfaces(h, &wg)
				<-sem
			}(hostname)
		}

		wg.Wait()
		fmt.Println("Interface config collection completed.")
	},
}

func isRouter(hostname string) bool {
	// Check for common router patterns
	routerPatterns := []string{
		"AS1", "AS2", "AS3", "AS4",  // ASR routers
		"IR1", "IR2", "IR3",         // Internal routers  
		"RR1", "RR2",                // Route reflectors
		"ASR",                       // Explicit ASR
		"CE1", "CE2",                // Customer Edge
	}
	
	upperHostname := strings.ToUpper(hostname)
	for _, pattern := range routerPatterns {
		if strings.Contains(upperHostname, pattern) {
			return true
		}
	}
	
	// Exclude switches and management devices
	if strings.Contains(upperHostname, "SW") || strings.Contains(upperHostname, "MS") {
		return false
	}
	
	return true // Default to true for unknown patterns
}

func collectDeviceInterfaces(hostname string, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Collecting interfaces from %s...\n", hostname)

	result, err := client.SendDeviceCommand(hostname, interfacesCommand, true, true)
	if err != nil {
		fmt.Printf("Error collecting interfaces from %s: %v\n", hostname, err)
		return
	}

	filename := filepath.Join(interfacesOutputDir, hostname+"_interfaces.txt")
	err = os.WriteFile(filename, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Error writing file for %s: %v\n", hostname, err)
		return
	}

	fmt.Printf("Saved interface config for %s to %s\n", hostname, filename)
}

func init() {
	rootCmd.AddCommand(collectInterfacesCmd)

	// POSIX-style flags
	collectInterfacesCmd.Flags().StringVarP(&interfacesOutputDir, "output-dir", "o", "interfaces", "Output directory for interface config files")
	collectInterfacesCmd.Flags().IntVarP(&interfacesConcurrency, "concurrency", "c", 10, "Number of concurrent workers")
	collectInterfacesCmd.Flags().BoolVarP(&interfacesFilterRouters, "routers-only", "r", true, "Only collect from routers (filter out switches)")
	collectInterfacesCmd.Flags().StringVar(&interfacesCommand, "command", "show running-config | section interface", "Command to collect interface configs")

	// Validation
	collectInterfacesCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if interfacesConcurrency < 1 {
			return fmt.Errorf("concurrency must be at least 1, got %d", interfacesConcurrency)
		}
		return nil
	}
}