package main

import (
	"encoding/json"
	"fmt"
	"os"

	"tomclient/internal"
	"tomclient/tomapi"
)

func sendDeviceCommand(client *tomapi.Client, deviceName, command string) {
	result, err := client.SendDeviceCommand(deviceName, command, true, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(result)
}

func exportInventory(client *tomapi.Client, filter string) {
	inventory, err := client.ExportInventory(filter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	prettyJSON, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(prettyJSON))
}



func main() {
	baseURL := os.Getenv("TOM_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	
	client := tomapi.NewClient(baseURL)
	
	args := os.Args[1:]
	
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}
	
	if args[0] == "export" {
		filter := ""
		if len(args) > 1 {
			filter = args[1]
		}
		exportInventory(client, filter)
		return
	}
	
	if args[0] == "cmd" && len(args) >= 3 {
		deviceName := args[1]
		command := args[2]
		sendDeviceCommand(client, deviceName, command)
		return
	}
	
	if args[0] == "bulk-inventory" && len(args) >= 2 {
		devicesFile := args[1]
		concurrency := 20
		
		if len(args) >= 3 {
			parsed, err := internal.ParseConcurrency(args[2])
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
			concurrency = parsed
		}
		
		err := internal.BulkInventory(client, devicesFile, concurrency)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	
	if args[0] == "report" {
		err := internal.GenerateInventoryReport("inventory")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Inventory report generated: inventory_report.csv")
		return
	}
	
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  tomclient export [filter]       - export inventory (optionally filtered)")
	fmt.Println("  tomclient cmd <device> <cmd>    - run command on device")
	fmt.Println("  tomclient bulk-inventory <file> [concurrency] - run 'show inventory | i ASR' on all devices (default: 20 concurrent)")
	fmt.Println("  tomclient report                - generate CSV report from inventory files")
}