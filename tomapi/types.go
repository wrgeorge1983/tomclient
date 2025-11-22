package tomapi

import "time"

// JobResponse represents the response from job-based API calls
type JobResponse struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// DeviceConfig represents a device configuration from inventory
type DeviceConfig struct {
	Adapter        string                 `json:"adapter"`
	AdapterDriver  string                 `json:"adapter_driver"`
	AdapterOptions map[string]interface{} `json:"adapter_options"`
	Host           string                 `json:"host"`
	Port           int                    `json:"port"`
	CredentialID   string                 `json:"credential_id"`
}

// RawInventoryNode represents a raw inventory node (SolarWinds format)
type RawInventoryNode struct {
	NodeID      int    `json:"NodeID"`
	Caption     string `json:"Caption"`
	IPAddress   string `json:"IPAddress"`
	Vendor      string `json:"Vendor"`
	Description string `json:"Description"`
	Status      int    `json:"Status"`
	URI         string `json:"Uri"`
	DetailsURL  string `json:"DetailsUrl"`
}
