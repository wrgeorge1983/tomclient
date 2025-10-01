package internal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// InterfaceInfo represents an interface with its configuration
type InterfaceInfo struct {
	Name        string
	Description string
	Config      []string
}

// DeviceInterfaceInfo represents all interfaces for a device
type DeviceInterfaceInfo struct {
	Hostname   string
	Interfaces []InterfaceInfo
}

// ParseInterfaceConfig parses interface configuration from a file
func ParseInterfaceConfig(filename string) (*DeviceInterfaceInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	// Extract hostname from filename (remove _interfaces.txt)
	hostname := strings.TrimSuffix(strings.TrimSuffix(filename, "_interfaces.txt"), ".txt")
	if idx := strings.LastIndex(hostname, "/"); idx >= 0 {
		hostname = hostname[idx+1:]
	}
	if idx := strings.LastIndex(hostname, "\\"); idx >= 0 {
		hostname = hostname[idx+1:]
	}

	deviceInfo := &DeviceInterfaceInfo{
		Hostname:   hostname,
		Interfaces: []InterfaceInfo{},
	}

	scanner := bufio.NewScanner(file)
	var currentInterface *InterfaceInfo
	
	// Regex patterns
	interfacePattern := regexp.MustCompile(`^interface\s+(.+)$`)
	descriptionPattern := regexp.MustCompile(`^\s*description\s+(.+)$`)
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Check if this is an interface line
		if matches := interfacePattern.FindStringSubmatch(line); matches != nil {
			// Save previous interface if exists
			if currentInterface != nil {
				deviceInfo.Interfaces = append(deviceInfo.Interfaces, *currentInterface)
			}
			
			// Start new interface
			currentInterface = &InterfaceInfo{
				Name:   matches[1],
				Config: []string{line},
			}
		} else if currentInterface != nil {
			// Add line to current interface config
			currentInterface.Config = append(currentInterface.Config, line)
			
			// Check for description
			if matches := descriptionPattern.FindStringSubmatch(line); matches != nil {
				currentInterface.Description = matches[1]
			}
			
			// Check if we're at the end of this interface (next interface or end of config)
			if strings.HasPrefix(line, "!") || 
			   (len(strings.TrimSpace(line)) == 0 && len(currentInterface.Config) > 1) {
				// This might be the end of the interface block
				continue
			}
		}
	}
	
	// Don't forget the last interface
	if currentInterface != nil {
		deviceInfo.Interfaces = append(deviceInfo.Interfaces, *currentInterface)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	return deviceInfo, nil
}

// FindSSNInterfaces finds all interfaces with 'SSN' in the description
func FindSSNInterfaces(deviceInfo *DeviceInterfaceInfo) []InterfaceInfo {
	var ssnInterfaces []InterfaceInfo
	
	for _, iface := range deviceInfo.Interfaces {
		if strings.Contains(strings.ToUpper(iface.Description), "SSN") {
			ssnInterfaces = append(ssnInterfaces, iface)
		}
	}
	
	return ssnInterfaces
}

// isSubInterface determines if an interface is a subinterface (contains a dot)
func isSubInterface(interfaceName string) bool {
	return strings.Contains(interfaceName, ".")
}

// GenerateDeleteCommands generates Cisco commands to delete interfaces
func GenerateDeleteCommands(interfaces []InterfaceInfo) []string {
	var commands []string
	
	// Add configuration mode entry
	commands = append(commands, "configure terminal")
	
	for _, iface := range interfaces {
		// Use different commands based on interface type
		if isSubInterface(iface.Name) {
			commands = append(commands, fmt.Sprintf("no interface %s", iface.Name))
		} else {
			commands = append(commands, fmt.Sprintf("default interface %s", iface.Name))
		}
	}
	
	// Add exit and save commands
	commands = append(commands, "exit")
	commands = append(commands, "write memory")
	
	return commands
}

// GenerateDeleteCommandsDetailed generates detailed deletion commands with confirmation
func GenerateDeleteCommandsDetailed(interfaces []InterfaceInfo) []string {
	var commands []string
	
	// Add header comment
	commands = append(commands, "! Generated interface deletion commands")
	commands = append(commands, fmt.Sprintf("! Found %d interfaces with SSN in description", len(interfaces)))
	commands = append(commands, "!")
	commands = append(commands, "! WARNING: These commands will DELETE/RESET interfaces - REVIEW CAREFULLY")
	commands = append(commands, "! Subinterfaces: 'no interface X.Y' (removes subinterface)")
	commands = append(commands, "! Physical interfaces: 'default interface X' (resets to factory defaults)")
	commands = append(commands, "!")
	
	// Add configuration mode entry
	commands = append(commands, "configure terminal")
	commands = append(commands, "!")
	
	for _, iface := range interfaces {
		// Add comment showing what we're doing
		if isSubInterface(iface.Name) {
			commands = append(commands, fmt.Sprintf("! Removing subinterface %s - Description: %s", iface.Name, iface.Description))
			commands = append(commands, fmt.Sprintf("no interface %s", iface.Name))
		} else {
			commands = append(commands, fmt.Sprintf("! Resetting physical interface %s - Description: %s", iface.Name, iface.Description))
			commands = append(commands, fmt.Sprintf("default interface %s", iface.Name))
		}
		commands = append(commands, "!")
	}
	
	// Add exit and save commands
	commands = append(commands, "exit")
	commands = append(commands, "!")
	commands = append(commands, "! Save configuration")
	commands = append(commands, "write memory")
	
	return commands
}