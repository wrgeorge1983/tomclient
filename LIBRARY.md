# Using tomclient as a Library

The `tomapi` package provides a lightweight Go client for the Tom API that can be imported into other projects without pulling in CLI dependencies like OAuth flows, config files, or command-line parsing.

## Installation

```bash
go get github.com/wrgeorge1983/tomclient/tomapi
```

## Quick Start

### With API Key Authentication

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/wrgeorge1983/tomclient/tomapi"
)

func main() {
    // Create client with API key
    client := tomapi.NewClientWithAPIKey(
        "https://tom.example.com",
        "your-api-key-here",
    )
    
    // Get device inventory
    devices, err := client.GetInventory("")
    if err != nil {
        log.Fatalf("Failed to get inventory: %v", err)
    }
    
    for _, device := range devices {
        fmt.Printf("Device: %s (%s)\n", device.Name, device.IPAddress)
    }
}
```

### With Bearer Token Authentication

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/wrgeorge1983/tomclient/tomapi"
)

func main() {
    // Create client with bearer token (e.g., JWT from your own OAuth flow)
    client := tomapi.NewClientWithToken(
        "https://tom.example.com",
        "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    )
    
    // Execute command on a device
    result, err := client.ExecuteDeviceCommand("router1", "show version", 30, true, true)
    if err != nil {
        log.Fatalf("Failed to execute command: %v", err)
    }
    
    fmt.Printf("Command output:\n%s\n", result.Output)
}
```

### With No Authentication

```go
package main

import (
    "github.com/wrgeorge1983/tomclient/tomapi"
)

func main() {
    // Create client with no authentication
    client := tomapi.NewClient("http://localhost:8020", nil)
    
    devices, _ := client.GetInventory("")
    // ... use devices
}
```

## Advanced: Custom Authentication

Implement the `AuthProvider` interface for custom authentication schemes:

```go
package main

import (
    "fmt"
    "net/http"
    
    "github.com/wrgeorge1983/tomclient/tomapi"
)

// CustomAuth implements a custom authentication scheme
type CustomAuth struct {
    Username string
    Password string
}

func (a *CustomAuth) AddAuth(req *http.Request) error {
    // Add custom authentication headers
    req.Header.Set("X-Custom-User", a.Username)
    req.Header.Set("X-Custom-Pass", a.Password)
    return nil
}

func main() {
    // Use custom auth provider
    auth := &CustomAuth{
        Username: "admin",
        Password: "secret",
    }
    
    client := tomapi.NewClient("https://tom.example.com", auth)
    
    // Use client normally
    devices, err := client.GetInventory("")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d devices\n", len(devices))
}
```

## API Reference

### Client Creation

```go
// Create client with custom auth provider
func NewClient(baseURL string, authProvider AuthProvider) *Client

// Create client with API key (convenience)
func NewClientWithAPIKey(baseURL, apiKey string) *Client

// Create client with bearer token (convenience)
func NewClientWithToken(baseURL, token string) *Client
```

### Client Methods

```go
// Get device inventory (optionally filtered)
func (c *Client) GetInventory(filter string) ([]Device, error)

// Execute command on a device
func (c *Client) ExecuteDeviceCommand(
    deviceName string,
    command string,
    timeout int,
    wait bool,
    raw bool,
) (*CommandResult, error)
```

### AuthProvider Interface

```go
type AuthProvider interface {
    // AddAuth adds authentication to an HTTP request
    AddAuth(req *http.Request) error
}
```

### Built-in Auth Providers

```go
// No authentication
type NoAuth struct{}

// API key authentication
type APIKeyAuth struct {
    APIHeader string  // e.g., "X-API-Key"
    APIKey    string
}

// Bearer token authentication
type BearerTokenAuth struct {
    Token string
}
```

### Types

```go
type Device struct {
    Name      string `json:"name"`
    IPAddress string `json:"ip_address"`
    DeviceType string `json:"device_type"`
    // ... other fields
}

type CommandResult struct {
    Output    string `json:"output"`
    ExitCode  int    `json:"exit_code"`
    Error     string `json:"error,omitempty"`
    // ... other fields
}
```

## Key Benefits

- **No CLI Dependencies**: Import just the API client without OAuth libraries, config parsers, or CLI frameworks
- **Lightweight**: Minimal dependencies (standard library only)
- **Flexible Auth**: Use simple API keys, bearer tokens, or implement custom authentication
- **Type Safe**: Full Go type definitions for all API requests and responses

## Comparison: Library vs CLI

### Use the Library when:
- Building another application or service
- Integrating Tom API into existing Go code
- Want minimal dependencies
- Need programmatic control over authentication

### Use the CLI when:
- Interactive terminal usage
- Need OAuth/OIDC authentication flows
- Want config file management
- Using shell scripts

## Example: Automated Backup Tool

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/wrgeorge1983/tomclient/tomapi"
)

func main() {
    client := tomapi.NewClientWithAPIKey(
        os.Getenv("TOM_URL"),
        os.Getenv("TOM_API_KEY"),
    )
    
    // Get all devices
    devices, err := client.GetInventory("")
    if err != nil {
        log.Fatal(err)
    }
    
    // Backup configuration from each device
    for _, device := range devices {
        fmt.Printf("Backing up %s...\n", device.Name)
        
        result, err := client.ExecuteDeviceCommand(
            device.Name,
            "show running-config",
            60,
            true,
            true,
        )
        if err != nil {
            log.Printf("Failed to backup %s: %v", device.Name, err)
            continue
        }
        
        // Save to file
        filename := fmt.Sprintf("backups/%s-%s.cfg", 
            device.Name, 
            time.Now().Format("2006-01-02"),
        )
        os.WriteFile(filename, []byte(result.Output), 0644)
    }
}
```
