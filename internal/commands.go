package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"tomclient/tomapi"
)

// BulkInventory processes devices for inventory collection
func BulkInventory(client *tomapi.Client, devicesFile string, concurrency int) error {
	data, err := os.ReadFile(devicesFile)
	if err != nil {
		return fmt.Errorf("error reading devices file: %w", err)
	}
	
	var devices map[string]any
	err = json.Unmarshal(data, &devices)
	if err != nil {
		return fmt.Errorf("error parsing devices JSON: %w", err)
	}
	
	err = os.MkdirAll("inventory", 0755)
	if err != nil {
		return fmt.Errorf("error creating inventory directory: %w", err)
	}
	
	hostnames := make([]string, 0, len(devices))
	for hostname := range devices {
		hostnames = append(hostnames, hostname)
	}
	
	fmt.Printf("Processing %d devices with %d concurrent workers...\n", len(hostnames), concurrency)
	
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	
	for _, hostname := range hostnames {
		wg.Add(1)
		go func(h string) {
			sem <- struct{}{}
			processDevice(client, h, &wg)
			<-sem
		}(hostname)
	}
	
	wg.Wait()
	fmt.Println("All devices processed.")
	return nil
}

func processDevice(client *tomapi.Client, hostname string, wg *sync.WaitGroup) {
	defer wg.Done()
	
	fmt.Printf("Fetching inventory for %s...\n", hostname)
	
	result, err := client.SendDeviceCommand(hostname, "show inventory | i ASR", true, true)
	if err != nil {
		fmt.Printf("Error fetching inventory for %s: %v\n", hostname, err)
		return
	}
	
	filename := filepath.Join("inventory", hostname+"_inventory.txt")
	err = os.WriteFile(filename, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Error writing file for %s: %v\n", hostname, err)
		return
	}
	
	fmt.Printf("Saved inventory for %s to %s\n", hostname, filename)
}

// ParseConcurrency parses and validates concurrency argument
func ParseConcurrency(arg string) (int, error) {
	parsed, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid concurrency value: %s", arg)
	}
	if parsed < 1 {
		return 0, fmt.Errorf("concurrency must be at least 1")
	}
	return parsed, nil
}