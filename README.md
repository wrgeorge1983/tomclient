# tomclient

A CLI client for the Tom Smykowski network automation broker service. Execute commands on network devices, manage inventory, and perform bulk operations with support for OAuth, API key, or no authentication.

## Features

- **Device command execution** - Run commands on network devices via inventory
- **Bulk operations** - Execute commands across multiple devices concurrently
- **Inventory management** - Export and manage device inventory
- **Multiple auth modes** - OAuth (JWT), API key, or no authentication
- **Interface parsing** - Parse and analyze Cisco interface configurations
- **Hardware inventory** - Extract serial numbers and calculate device age

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd tomclient

# Build the binary
go build

# Install (optional)
go install
```

## Quick Start

### Basic Usage (No Auth)

```bash
# Set API URL
export TOM_API_URL=http://localhost:8020

# Run a command on a device
tomclient device router1 "show version"

# Get inventory
tomclient export inventory

# Bulk inventory collection
tomclient bulk-inventory devices.json --concurrency=20
```

### With OAuth Authentication

```bash
# Configure OAuth (works with any OIDC provider)
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_CLIENT_ID=your-client-id
export TOM_OAUTH_DISCOVERY_URL=https://accounts.google.com/.well-known/openid-configuration

# Authenticate (opens browser)
tomclient auth login

# Check auth status
tomclient auth status

# Run commands (uses stored token)
tomclient device router1 "show version"
```

### With API Key Authentication

```bash
# Configure API key
export TOM_AUTH_MODE=api_key
export TOM_API_KEY=your-secret-api-key

# Run commands
tomclient device router1 "show version"
```

## Commands

### Authentication

```bash
# Authenticate with OAuth
tomclient auth login

# Check authentication status
tomclient auth status

# Logout (clear stored token)
tomclient auth logout
```

### Device Operations

```bash
# Execute command on a device
tomclient device <device-name> <command> [flags]

# Options:
#   -t, --timeout int      Command timeout in seconds (default 10)
#   -w, --wait            Wait for command completion (default true)
#   -r, --raw             Return raw command output (default true)
#   -u, --username string Override username for authentication
#   -p, --password string Override password for authentication

# Examples:
tomclient device router1 "show version"
tomclient device switch1 "show interface" --timeout=30
tomclient device router2 "show running-config" -u admin -p secret
```

### Bulk Operations

```bash
# Run inventory command on multiple devices
tomclient bulk-inventory <devices-file> [flags]

# Options:
#   -c, --concurrency int  Number of concurrent workers (default 20)
#   -o, --output-dir string Output directory for inventory files (default "inventory")

# Example:
tomclient bulk-inventory devices.json --concurrency=10 --output-dir=./data
```

### Inventory Management

```bash
# Export inventory
tomclient export inventory [flags]

# Generate inventory report
tomclient report [flags]
#   -i, --input-dir string  Directory containing inventory files (default "inventory")
#   -o, --output string     Output CSV file name (default "inventory_report.csv")
```

### Interface Management

```bash
# Parse interface configurations
tomclient parse-interfaces <file>

# Collect interfaces from devices
tomclient collect-interfaces [flags]
```

## Configuration

### Authentication Modes

tomclient supports three authentication modes:

1. **`none`** (default) - No authentication
2. **`api_key`** - Static API key authentication
3. **`jwt`** - OAuth/PKCE authentication with Duo or other OIDC providers

Set the mode via `TOM_AUTH_MODE` environment variable or config file.

### Configuration Hierarchy

Configuration is loaded in this order (later sources override earlier):

1. **Config file** - `~/.tom/config.json` (or path specified by `--config-dir`)
2. **Environment variables** - `TOM_*` variables
3. **Command-line flags** - e.g., `--api-url`, `--config-dir`

**Example:** If `api_url` is set in the config file, `TOM_API_URL` environment variable, and `--api-url` flag, the flag value wins.

### Config File Location

Default: `~/.tom/config.json`

Override with:
- `--config-dir` flag: `tomclient --config-dir=/path/to/config`
- Environment variable: `export TOM_CONFIG_DIR=/path/to/config`

### Config File Format

All configuration options can be set in the config file:

```json
{
  "api_url": "http://localhost:8020",
  "auth_mode": "jwt",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration",
  "oauth_redirect_port": 8899,
  "oauth_scopes": "openid email profile"
}
```

For API key mode:
```json
{
  "api_url": "http://localhost:8020",
  "auth_mode": "api_key",
  "api_key": "your-secret-api-key"
}
```

### Environment Variables

| Variable | Required For | Default | Description |
|----------|-------------|---------|-------------|
| `TOM_API_URL` | All modes | `http://localhost:8000` | Tom API server URL |
| `TOM_AUTH_MODE` | - | `none` | Auth mode: `none`, `api_key`, or `jwt` |
| `TOM_API_KEY` | `api_key` mode | - | API key for authentication |
| `TOM_OAUTH_CLIENT_ID` | `jwt` mode | - | OAuth client ID |
| `TOM_OAUTH_DISCOVERY_URL` | `jwt` mode | - | Full URL to OIDC discovery document |
| `TOM_OAUTH_REDIRECT_PORT` | - | `8899` | Local callback server port for OAuth |
| `TOM_OAUTH_SCOPES` | - | `openid email profile` | OAuth scopes to request |
| `TOM_CONFIG_DIR` | - | `~/.tom` | Config directory path |

### Global Flags

```bash
-a, --api-url string      Tom API server URL (default "http://localhost:8000")
    --config-dir string   Config directory path (default "~/.tom")
-h, --help               Help for any command
```

## Authentication Details

### OAuth (JWT Mode)

When using `auth_mode=jwt`, tomclient uses OAuth 2.0 with PKCE (Proof Key for Code Exchange) for secure authentication without client secrets. You provide the full OIDC discovery URL, and the client fetches the provider's OAuth endpoints.

**Flow:**
1. Run `tomclient auth login`
2. Client fetches your provider's OIDC discovery document from `TOM_OAUTH_DISCOVERY_URL`
3. Browser opens to the discovered authorization endpoint
4. Complete authentication (MFA, SSO, etc.)
5. Browser redirects to localhost callback
6. Token is exchanged at the discovered token endpoint and saved to `~/.tom/token.json` with 0600 permissions
7. Subsequent commands use the stored token automatically

**Finding your discovery URL:** Most OIDC providers have their discovery document at `{base_url}/.well-known/openid-configuration`. Check your provider's documentation or try accessing that URL in your browser - it should return a JSON document with endpoints.

**Token management:**
- Tokens expire after the period specified by your IdP (typically 1 hour)
- Tokens are automatically validated before each request
- If expired, you'll see: `Error: token expired - run 'tomclient auth login'`
- Tokens are stored with 60-second expiration buffer for safety

**Manual URL fallback:**
If the browser doesn't open automatically (e.g., SSH session), the authentication URL is printed for manual copy/paste.

### API Key Mode

Simple static key authentication. The API key is sent in the `X-API-Key` header with every request.

```bash
export TOM_AUTH_MODE=api_key
export TOM_API_KEY=your-secret-key
tomclient device router1 "show version"
```

### No Authentication Mode

For testing or when the Tom API is running with `TOM_CORE_AUTH_MODE=none`.

```bash
export TOM_AUTH_MODE=none
tomclient device router1 "show version"
```

## Configuration Validation

tomclient validates configuration on startup and provides clear error messages:

**Missing required config:**
```
Error: auth_mode is 'jwt' but TOM_DUO_CLIENT_ID is not set
```

**Unused configuration warnings:**
```
Warning: TOM_API_KEY is set but auth_mode is 'none' - API key will not be used
```

**Not authenticated:**
```
Error: not authenticated - run 'tomclient auth login' first
```

## Development & Testing

### Using a Test Config Directory

For development or testing, use a local config directory:

```bash
# Create test config
mkdir -p .tom-test
cat > .tom-test/config.json << 'EOF'
{
  "api_url": "http://localhost:8020",
  "auth_mode": "jwt",
  "oauth_client_id": "test-client-id",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration"
}
EOF

# Use it
tomclient --config-dir=.tom-test auth status
tomclient --config-dir=.tom-test device router1 "show version"
```

### Building

```bash
go build
./tomclient --version
```

### Running Tests

```bash
go test ./...
```

## Examples

### Device Command Execution

```bash
# Basic command
tomclient device ASHBVA21AS1 "show version"

# With timeout
tomclient device ASHBVA21AS1 "show inventory" --timeout=30

# With credential override
tomclient device ASHBVA21AS1 "show running-config" -u admin -p secret
```

### Bulk Inventory Collection

```bash
# Create devices file
cat > devices.json << 'EOF'
{
  "ASHBVA21AS1": "192.168.1.1",
  "ASHBVA21AS2": "192.168.1.2",
  "ATLAGA18AS1": "192.168.2.1"
}
EOF

# Run bulk collection
tomclient bulk-inventory devices.json --concurrency=20 --output-dir=inventory

# Generate report from collected data
tomclient report --input-dir=inventory --output=devices.csv
```

### Inventory Management

```bash
# Export all devices
tomclient export inventory

# Export with filter
tomclient export inventory --filter=routers

# List available filters
tomclient export filters
```

### Interface Analysis

```bash
# Parse interface config from file
tomclient parse-interfaces interfaces/ASHBVA21AS1_interfaces.txt

# Collect interfaces from multiple devices
tomclient collect-interfaces --devices=routers.json
```

## File Structure

```
tomclient/
├── auth/              # Authentication package
│   ├── config.go      # Config management
│   ├── oauth.go       # OAuth/PKCE flow
│   ├── pkce.go        # PKCE code generation
│   └── token.go       # Token storage
├── cmd/               # CLI commands
│   ├── root.go        # Root command
│   ├── auth.go        # Auth subcommands
│   ├── device.go      # Device operations
│   ├── bulk.go        # Bulk operations
│   ├── export.go      # Inventory export
│   └── report.go      # Report generation
├── internal/          # Internal packages
│   ├── commands.go    # Command utilities
│   ├── parser.go      # Inventory parsing
│   ├── interface_parser.go  # Interface parsing
│   └── reports.go     # Report generation
├── tomapi/            # Tom API client
│   ├── client.go      # HTTP client
│   ├── devices.go     # Device methods
│   ├── inventory.go   # Inventory methods
│   └── types.go       # Data types
└── main.go            # Entry point
```

## Error Handling

tomclient provides clear, actionable error messages:

| Error | Meaning | Solution |
|-------|---------|----------|
| `auth_mode is 'jwt' but TOM_OAUTH_CLIENT_ID is not set` | Missing OAuth config | Set `TOM_OAUTH_CLIENT_ID` and `TOM_OAUTH_DISCOVERY_URL` |
| `failed to fetch OIDC discovery from ...` | Can't fetch discovery document | Check `TOM_OAUTH_DISCOVERY_URL` is correct and accessible |
| `not authenticated - run 'tomclient auth login' first` | No stored token | Run `tomclient auth login` |
| `token expired - run 'tomclient auth login'` | Token expired | Re-authenticate with `tomclient auth login` |
| `API returned status code: 401` | Invalid/missing auth | Check auth mode and credentials |
| `failed to load config` | Config file issue | Check `~/.tom/config.json` syntax |

## Security

- **Token files** (`~/.tom/token.json`) stored with 0600 permissions
- **Config files** (`~/.tom/config.json`) stored with 0600 permissions
- **PKCE flow** prevents authorization code interception
- **State parameter** prevents CSRF attacks during OAuth
- **60-second expiration buffer** prevents using expired tokens
- **No token logging** - Sensitive data never printed to console

## Troubleshooting

### OAuth browser doesn't open

The authentication URL is always printed. Copy and paste it into your browser manually:

```
If the browser doesn't open automatically, visit this URL:
https://sso-xxxxx.sso.duosecurity.com/oidc/...
```

### Port 8899 already in use

The OAuth callback server will automatically find a free port and update the redirect URI accordingly.

### Token expired

Tokens expire after the period set by your identity provider (typically 1 hour). Simply re-authenticate:

```bash
tomclient auth login
```

### Config not loading

Check the config directory being used:

```bash
tomclient auth status
# Shows: Config Dir: /home/user/.tom
```

Override if needed:
```bash
tomclient --config-dir=/custom/path auth status
```

## API Reference

See [api-endpoints.md](api-endpoints.md) for Tom API endpoint documentation.

See [AUTH.md](AUTH.md) for detailed authentication guide.

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]
