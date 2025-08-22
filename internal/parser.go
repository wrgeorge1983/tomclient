package internal

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ParseInventoryFile parses an inventory file and extracts serial numbers
func ParseInventoryFile(filename string) (chassis, rp, esp, allSerials []string) {
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

// CalculateAge calculates the age of a device based on its serial number
func CalculateAge(serialNumber string) int {
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

// CalculateAverageAge calculates the average age from a list of serial numbers
func CalculateAverageAge(serials []string) float64 {
	if len(serials) == 0 {
		return 0
	}
	
	totalAge := 0
	validSerials := 0
	
	for _, serial := range serials {
		age := CalculateAge(serial)
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