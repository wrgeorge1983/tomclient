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

// SendCommandRequest is the request body for POST /device/{name}/send_command
type SendCommandRequest struct {
	Command      string  `json:"command"`
	Wait         bool    `json:"wait,omitempty"`
	Parse        bool    `json:"parse,omitempty"`
	Parser       string  `json:"parser,omitempty"`
	Template     *string `json:"template,omitempty"`
	IncludeRaw   bool    `json:"include_raw,omitempty"`
	Timeout      int     `json:"timeout,omitempty"`
	UseCache     bool    `json:"use_cache,omitempty"`
	CacheTTL     *int    `json:"cache_ttl,omitempty"`
	CacheRefresh bool    `json:"cache_refresh,omitempty"`
	Username     *string `json:"username,omitempty"`
	Password     *string `json:"password,omitempty"`
}

// CommandSpec specifies a single command with optional parsing configuration
type CommandSpec struct {
	Command    string  `json:"command"`
	Parse      *bool   `json:"parse,omitempty"`
	Parser     *string `json:"parser,omitempty"`
	Template   *string `json:"template,omitempty"`
	IncludeRaw *bool   `json:"include_raw,omitempty"`
}

// SendCommandsRequest is the request body for POST /device/{name}/send_commands
type SendCommandsRequest struct {
	Commands     []interface{} `json:"commands"` // Can be []string or []CommandSpec
	Parse        bool          `json:"parse,omitempty"`
	Parser       string        `json:"parser,omitempty"`
	IncludeRaw   bool          `json:"include_raw,omitempty"`
	Wait         bool          `json:"wait,omitempty"`
	Timeout      int           `json:"timeout,omitempty"`
	Retries      int           `json:"retries,omitempty"`
	MaxQueueWait int           `json:"max_queue_wait,omitempty"`
	UseCache     bool          `json:"use_cache,omitempty"`
	CacheRefresh bool          `json:"cache_refresh,omitempty"`
	CacheTTL     *int          `json:"cache_ttl,omitempty"`
	Username     *string       `json:"username,omitempty"`
	Password     *string       `json:"password,omitempty"`
}

// RawCommandRequest is the request body for POST /raw/send_netmiko_command and /raw/send_scrapli_command
type RawCommandRequest struct {
	Host         string  `json:"host"`
	DeviceType   string  `json:"device_type"`
	Command      string  `json:"command"`
	Port         int     `json:"port,omitempty"`
	Wait         bool    `json:"wait,omitempty"`
	CredentialID *string `json:"credential_id,omitempty"`
	Username     *string `json:"username,omitempty"`
	Password     *string `json:"password,omitempty"`
}
