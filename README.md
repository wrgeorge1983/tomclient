# tomclient

A CLI client for the Tom Smykowski network automation broker service.

Authenticate with OAuth, API keys, or no auth, then execute commands on network devices or export inventory.

## Installation

```bash
go build
```

## Quick Start

### With No Authentication

```bash
export TOM_API_URL=http://localhost:8020
export TOM_AUTH_MODE=none

tomclient device router1 "show version"
```

### With API Key

```bash
export TOM_API_URL=http://localhost:8020
export TOM_AUTH_MODE=api_key
export TOM_API_KEY=your-secret-key

tomclient device router1 "show version"
```

### With OAuth

```bash
export TOM_API_URL=http://localhost:8020
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_CLIENT_ID=your-client-id
export TOM_OAUTH_DISCOVERY_URL=https://accounts.google.com/.well-known/openid-configuration

# Authenticate (opens browser)
tomclient auth login

# Run commands
tomclient device router1 "show version"
```

## Commands

### Authentication

```bash
# Login with OAuth
tomclient auth login

# Check authentication status
tomclient auth status

# Logout (clear stored token)
tomclient auth logout
```

### Device Commands

Execute commands on network devices:

```bash
tomclient device <device-name> <command> [flags]

# Examples
tomclient device router1 "show version"
tomclient device switch1 "show interface" --timeout=30
tomclient device router2 "show ip route" --raw
```

**Flags:**
- `-t, --timeout int` - Command timeout in seconds (default: 10)
- `-w, --wait` - Wait for command completion (default: true)
- `-r, --raw` - Return raw output (default: true)
- `-u, --username string` - Override username
- `-p, --password string` - Override password

### Export Inventory

Export device inventory from Tom API:

```bash
tomclient export [flags]

# Examples
tomclient export                          # All devices, pretty JSON
tomclient export --filter=routers         # Filtered devices
tomclient export --output=json            # Compact JSON
```

**Flags:**
- `-f, --filter string` - Filter name (optional)
- `-o, --output string` - Output format: `json` or `pretty` (default: pretty)

## Configuration

### Config File

Store settings in `~/.tom/config.json`:

```json
{
  "api_url": "http://localhost:8020",
  "auth_mode": "jwt",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration"
}
```

### Override Config Location

```bash
# Via flag
tomclient --config-dir=/path/to/config auth status

# Via environment variable
export TOM_CONFIG_DIR=/path/to/config
tomclient auth status
```

### Configuration Precedence

Settings are loaded in this order (later overrides earlier):

1. Config file (`~/.tom/config.json`)
2. Environment variables (`TOM_*`)
3. Command-line flags

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TOM_API_URL` | `http://localhost:8000` | Tom API server URL |
| `TOM_AUTH_MODE` | `none` | Auth mode: `none`, `api_key`, or `jwt` |
| `TOM_API_KEY` | - | API key (for `api_key` mode) |
| `TOM_OAUTH_CLIENT_ID` | - | OAuth client ID (for `jwt` mode) |
| `TOM_OAUTH_DISCOVERY_URL` | - | OIDC discovery URL (for `jwt` mode) |
| `TOM_OAUTH_REDIRECT_PORT` | `8899` | OAuth callback port |
| `TOM_OAUTH_SCOPES` | `openid email profile` | OAuth scopes |
| `TOM_CONFIG_DIR` | `~/.tom` | Config directory path |

### Global Flags

Available on all commands:

- `-a, --api-url string` - Tom API server URL
- `--config-dir string` - Config directory path
- `-h, --help` - Help for command

## Authentication Modes

### None (Default)

No authentication. Use when Tom API runs with `TOM_CORE_AUTH_MODE=none`.

```bash
export TOM_AUTH_MODE=none
```

### API Key

Static key authentication. Key sent in `X-API-Key` header.

```bash
export TOM_AUTH_MODE=api_key
export TOM_API_KEY=your-secret-key
```

### OAuth (JWT)

OAuth 2.0 with PKCE. Works with any OIDC-compliant provider (Google, Microsoft, Duo, Okta, etc.).

```bash
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_CLIENT_ID=your-client-id
export TOM_OAUTH_DISCOVERY_URL=https://provider/.well-known/openid-configuration

tomclient auth login    # Authenticate via browser
```

Tokens stored in `~/.tom/token.json` and used automatically for API requests.

See [AUTH.md](AUTH.md) for detailed authentication documentation.

## Examples

### Basic Device Command

```bash
tomclient device ROUTER1 "show version"
```

### Device Command with Timeout

```bash
tomclient device ROUTER1 "show running-config" --timeout=60
```

### Export All Devices

```bash
tomclient export
```

### Export Filtered Devices

```bash
tomclient export --filter=routers --output=json
```

### Using Config File

```bash
# Create config
mkdir -p ~/.tom
cat > ~/.tom/config.json << EOF
{
  "api_url": "http://localhost:8020",
  "auth_mode": "api_key",
  "api_key": "my-secret-key"
}
EOF

# Use it
tomclient device router1 "show version"
```

### OAuth Authentication

```bash
# Configure
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_CLIENT_ID=abc123
export TOM_OAUTH_DISCOVERY_URL=https://accounts.google.com/.well-known/openid-configuration

# Login
tomclient auth login

# Check status
tomclient auth status

# Use authenticated commands
tomclient device router1 "show version"
tomclient export --filter=switches
```

## Error Handling

Common errors and solutions:

| Error | Solution |
|-------|----------|
| `Configuration error: auth_mode is 'jwt' but TOM_OAUTH_CLIENT_ID is not set` | Set required OAuth variables |
| `not authenticated - run 'tomclient auth login' first` | Run `tomclient auth login` |
| `token expired - run 'tomclient auth login'` | Re-authenticate |
| `API returned status code: 401` | Check auth mode and credentials |
| `failed to fetch OIDC discovery from ...` | Verify discovery URL is correct |

## Project Structure

```
tomclient/
├── auth/              # OAuth/PKCE authentication
├── cmd/               # CLI commands (auth, device, export)
├── tomapi/            # Tom API client library
├── AUTH.md            # Authentication guide
├── README.md          # This file
├── go.mod             # Go dependencies
└── main.go            # Entry point
```

## Development

### Build

```bash
go build
```

### Test with Local Config

```bash
mkdir -p .tom-test
cat > .tom-test/config.json << EOF
{
  "api_url": "http://localhost:8020",
  "auth_mode": "none"
}
EOF

./tomclient --config-dir=.tom-test device test-device "show version"
```

## License

[Add license here]
