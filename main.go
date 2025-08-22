package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func makeHTTPRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	
	apiKey := os.Getenv("TOM_API_KEY")
	if apiKey != "" {
		headerName := os.Getenv("TOM_API_KEY_HEADER")
		if headerName == "" {
			headerName = "X-API-Key"
		}
		req.Header.Set(headerName, apiKey)
	}
	
	client := &http.Client{}
	return client.Do(req)
}

func sendDeviceCommand(baseURL, deviceName, command string) {
	apiURL := fmt.Sprintf("%s/api/device/%s/send_command", baseURL, deviceName)
	
	params := url.Values{}
	params.Add("command", command)
	params.Add("wait", "true")
	params.Add("rawOutput", "true")
	
	fullURL := apiURL + "?" + params.Encode()
	
	resp, err := makeHTTPRequest("GET", fullURL)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API returned status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}
	
	var result string
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Print(string(body))
	} else {
		fmt.Print(result)
	}
}

func exportInventory(baseURL, filter string) {
	apiURL := baseURL + "/api/inventory/export"
	if filter != "" {
		apiURL += "?filter_name=" + url.QueryEscape(filter)
	}
	
	resp, err := makeHTTPRequest("GET", apiURL)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API returned status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}
	
	var inventory map[string]any
	err = json.Unmarshal(body, &inventory)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}
	
	prettyJSON, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(prettyJSON))
}

func processDevice(baseURL, hostname string, wg *sync.WaitGroup) {
	defer wg.Done()
	
	fmt.Printf("Fetching inventory for %s...\n", hostname)
	
	apiURL := fmt.Sprintf("%s/api/device/%s/send_command", baseURL, hostname)
	params := url.Values{}
	params.Add("command", "show inventory | i ASR")
	params.Add("wait", "true")
	params.Add("rawOutput", "true")
	
	fullURL := apiURL + "?" + params.Encode()
	
	resp, err := makeHTTPRequest("GET", fullURL)
	if err != nil {
		fmt.Printf("Error fetching inventory for %s: %v\n", hostname, err)
		return
	}
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API error for %s: status code %d\n", hostname, resp.StatusCode)
		resp.Body.Close()
		return
	}
	
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Printf("Error reading response for %s: %v\n", hostname, err)
		return
	}
	
	var result string
	err = json.Unmarshal(body, &result)
	if err != nil {
		result = string(body)
	}
	
	filename := filepath.Join("inventory", hostname+"_inventory.txt")
	err = os.WriteFile(filename, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Error writing file for %s: %v\n", hostname, err)
		return
	}
	
	fmt.Printf("Saved inventory for %s to %s\n", hostname, filename)
}

func bulkInventory(baseURL, devicesFile string, concurrency int) {
	data, err := os.ReadFile(devicesFile)
	if err != nil {
		fmt.Printf("Error reading devices file: %v\n", err)
		os.Exit(1)
	}
	
	var devices map[string]any
	err = json.Unmarshal(data, &devices)
	if err != nil {
		fmt.Printf("Error parsing devices JSON: %v\n", err)
		os.Exit(1)
	}
	
	err = os.MkdirAll("inventory", 0755)
	if err != nil {
		fmt.Printf("Error creating inventory directory: %v\n", err)
		os.Exit(1)
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
			processDevice(baseURL, h, &wg)
			<-sem
		}(hostname)
	}
	
	wg.Wait()
	fmt.Println("All devices processed.")
}

func calculateAge(serialNumber string) int {
	if len(serialNumber) < 5 {
		return -1
	}
	
	yearCodeStr := serialNumber[3:5]
	yearCode, err := strconv.Atoi(yearCodeStr)
	if err != nil {
		return -1
	}
	
	age := 25 - (yearCode - 4)
	return age
}

func parseInventoryFile(filename string) (chassis, rp, esp, allSerials []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, nil
	}
	
	content := string(data)
	lines := strings.Split(content, "\n")
	
	serialRegex := regexp.MustCompile(`SN: ([A-Z0-9]+)`)
	
	for i, line := range lines {
		if strings.Contains(line, "Chassis") && strings.Contains(line, "ASR") {
			if i+1 < len(lines) {
				matches := serialRegex.FindStringSubmatch(lines[i+1])
				if len(matches) > 1 {
					chassis = append(chassis, matches[1])
					allSerials = append(allSerials, matches[1])
				}
			}
		} else if strings.Contains(line, "Route Processor") {
			if i+1 < len(lines) {
				matches := serialRegex.FindStringSubmatch(lines[i+1])
				if len(matches) > 1 {
					rp = append(rp, matches[1])
					allSerials = append(allSerials, matches[1])
				}
			}
		} else if strings.Contains(line, "Embedded Services Processor") {
			if i+1 < len(lines) {
				matches := serialRegex.FindStringSubmatch(lines[i+1])
				if len(matches) > 1 {
					esp = append(esp, matches[1])
					allSerials = append(allSerials, matches[1])
				}
			}
		} else if strings.Contains(line, "SN:") {
			matches := serialRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				allSerials = append(allSerials, matches[1])
			}
		}
	}
	
	return chassis, rp, esp, allSerials
}

func calculateAverageAge(serials []string) float64 {
	if len(serials) == 0 {
		return 0
	}
	
	totalAge := 0
	validSerials := 0
	
	for _, serial := range serials {
		age := calculateAge(serial)
		if age >= 0 {
			totalAge += age
			validSerials++
		}
	}
	
	if validSerials == 0 {
		return 0
	}
	
	return float64(totalAge) / float64(validSerials)
}

func generateInventoryReport(inventoryDir string) {
	files, err := filepath.Glob(filepath.Join(inventoryDir, "*_inventory.txt"))
	if err != nil {
		fmt.Printf("Error finding inventory files: %v\n", err)
		os.Exit(1)
	}
	
	csvFile, err := os.Create("inventory_report.csv")
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		os.Exit(1)
	}
	defer csvFile.Close()
	
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()
	
	headers := []string{"Hostname", "Chassis_SN", "Chassis_Age", "RP1_SN", "RP1_Age", "RP2_SN", "RP2_Age", "ESP1_SN", "ESP1_Age", "ESP2_SN", "ESP2_Age", "Avg_Major_Age", "Avg_All_Age"}
	writer.Write(headers)
	
	for _, file := range files {
		basename := filepath.Base(file)
		hostname := strings.TrimSuffix(basename, "_inventory.txt")
		
		chassis, rp, esp, allSerials := parseInventoryFile(file)
		
		row := []string{hostname}
		
		if len(chassis) > 0 {
			row = append(row, chassis[0], strconv.Itoa(calculateAge(chassis[0])))
		} else {
			row = append(row, "", "")
		}
		
		for i := 0; i < 2; i++ {
			if i < len(rp) {
				row = append(row, rp[i], strconv.Itoa(calculateAge(rp[i])))
			} else {
				row = append(row, "", "")
			}
		}
		
		for i := 0; i < 2; i++ {
			if i < len(esp) {
				row = append(row, esp[i], strconv.Itoa(calculateAge(esp[i])))
			} else {
				row = append(row, "", "")
			}
		}
		
		majorSerials := make([]string, 0)
		majorSerials = append(majorSerials, chassis...)
		majorSerials = append(majorSerials, rp...)
		majorSerials = append(majorSerials, esp...)
		
		avgMajorAge := calculateAverageAge(majorSerials)
		avgAllAge := calculateAverageAge(allSerials)
		
		row = append(row, fmt.Sprintf("%.1f", avgMajorAge), fmt.Sprintf("%.1f", avgAllAge))
		
		writer.Write(row)
	}
	
	fmt.Println("Inventory report generated: inventory_report.csv")
}

func main() {
	baseURL := os.Getenv("TOM_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	
	args := os.Args[1:]
	
	if len(args) == 0 {
		fmt.Println("Usage:")
		fmt.Println("  tomclient export [filter]       - export inventory (optionally filtered)")
		fmt.Println("  tomclient cmd <device> <cmd>    - run command on device")
		fmt.Println("  tomclient bulk-inventory <file> [concurrency] - run 'show inventory | i ASR' on all devices (default: 20 concurrent)")
		fmt.Println("  tomclient report                - generate CSV report from inventory files")
		os.Exit(1)
	}
	
	if args[0] == "export" {
		filter := ""
		if len(args) > 1 {
			filter = args[1]
		}
		exportInventory(baseURL, filter)
		return
	}
	
	if args[0] == "cmd" && len(args) >= 3 {
		deviceName := args[1]
		command := args[2]
		sendDeviceCommand(baseURL, deviceName, command)
		return
	}
	
	if args[0] == "bulk-inventory" && len(args) >= 2 {
		devicesFile := args[1]
		concurrency := 20
		
		if len(args) >= 3 {
			parsed, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Printf("Invalid concurrency value: %s\n", args[2])
				os.Exit(1)
			}
			if parsed < 1 {
				fmt.Printf("Concurrency must be at least 1\n")
				os.Exit(1)
			}
			concurrency = parsed
		}
		
		bulkInventory(baseURL, devicesFile, concurrency)
		return
	}
	
	if args[0] == "report" {
		generateInventoryReport("inventory")
		return
	}
	
	fmt.Println("Usage:")
	fmt.Println("  tomclient export [filter]       - export inventory (optionally filtered)")
	fmt.Println("  tomclient cmd <device> <cmd>    - run command on device")
	fmt.Println("  tomclient bulk-inventory <file> [concurrency] - run 'show inventory | i ASR' on all devices (default: 20 concurrent)")
	fmt.Println("  tomclient report                - generate CSV report from inventory files")
	os.Exit(1)
}