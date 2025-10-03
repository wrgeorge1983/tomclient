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

**Using config.json (recommended):**
```bash
# Create ~/.tom/config.json
cat > ~/.tom/config.json << EOF
{
  "api_url": "http://localhost:8020",
  "auth_mode": "jwt",
  "oauth_provider": "google",
  "oauth_client_id": "xxx.apps.googleusercontent.com",
  "oauth_client_secret": "GOCSPX-xxxxx",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration"
}
EOF

# Authenticate (opens browser)
tomclient auth login

# Run commands
tomclient device router1 "show version"
```

**Or using environment variables:**
```bash
export TOM_API_URL=http://localhost:8020
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_PROVIDER=google
export TOM_OAUTH_CLIENT_ID=xxx.apps.googleusercontent.com
export TOM_OAUTH_CLIENT_SECRET=GOCSPX-xxxxx
export TOM_OAUTH_DISCOVERY_URL=https://accounts.google.com/.well-known/openid-configuration

tomclient auth login
tomclient device router1 "show version"
```

## Commands

### Inventory Cache

List and cache device hostnames for fast lookups and shell autocomplete:

```bash
# List all devices (auto-fetches and caches)
tomclient inventory

# Force refresh from API
tomclient inventory --refresh

# Filter by prefix
tomclient inventory --prefix=router

# Output in /etc/hosts format
tomclient inventory --hostfile

# Update /etc/hosts with device entries (requires sudo)
sudo tomclient inventory --update-hosts

# Clear cache
tomclient inventory clear
```

The inventory cache:
- Stored in `~/.tom/inventory_cache.json`
- Default TTL: 1 hour (configurable via `TOM_CACHE_TTL`)
- Automatically refreshed when expired
- Used for device name autocomplete

The `/etc/hosts` integration:
- Creates managed block in `/etc/hosts`
- Maps device names to IP addresses from inventory
- Updates are atomic (writes to temp file, then renames)
- Preserves existing entries outside managed block

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

### Hosts File Management

Automatically populate `/etc/hosts` with device names from inventory:

```bash
# Preview hostfile format
tomclient inventory --hostfile

# Update /etc/hosts (requires sudo)
sudo tomclient inventory --update-hosts

# Update with filtered devices
sudo tomclient inventory --prefix=prod --update-hosts
```

The managed block in `/etc/hosts` looks like:
```
# BEGIN tomclient managed block
# This section is automatically managed by tomclient
# Do not edit manually - changes will be overwritten
192.168.1.1    router1
192.168.1.2    router2
192.168.2.1    switch1
# END tomclient managed block
```

**Features:**
- Safe updates (atomic write via temp file)
- Preserves existing entries
- Clearly marked managed section
- Can be run repeatedly to update
- Sudo-aware: automatically uses your config/cache even when run with sudo

**Note:** When using `sudo`, tomclient automatically detects the original user and uses their config directory (`~/.tom/`), so authentication and cache work seamlessly.

## Configuration

### Config File (Recommended)

Store settings in `~/.tom/config.json`:

**Standard OIDC Provider (Duo, Okta, Keycloak, etc.):**
```json
{
  "api_url": "http://localhost:8020",
  "auth_mode": "jwt",
  "oauth_provider": "oidc",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://your-provider/.well-known/openid-configuration"
}
```

**Google Provider:**
```json
{
  "api_url": "http://localhost:8020",
  "auth_mode": "jwt",
  "oauth_provider": "google",
  "oauth_client_id": "xxx.apps.googleusercontent.com",
  "oauth_client_secret": "GOCSPX-xxxxx",
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

**Sudo Behavior:** When running with `sudo`, tomclient automatically detects the original user (via `SUDO_USER` environment variable) and uses their config directory. This means authentication and cached inventory work correctly even when elevated privileges are needed (e.g., for `--update-hosts`).

### Configuration Precedence

Settings are loaded in this order (later overrides earlier):

1. **Config file** (`~/.tom/config.json`) - Primary configuration
2. **Environment variables** (`TOM_*`) - Override config file
3. **Command-line flags** - Override both

### Environment Variables

Environment variables can override config file settings:

| Variable | Default | Description |
|----------|---------|-------------|
| `TOM_API_URL` | `http://localhost:8000` | Tom API server URL |
| `TOM_AUTH_MODE` | `none` | Auth mode: `none`, `api_key`, or `jwt` |
| `TOM_API_KEY` | - | API key (for `api_key` mode) |
| `TOM_OAUTH_PROVIDER` | `oidc` | OAuth provider: `oidc`, `google`, or `microsoft` |
| `TOM_OAUTH_CLIENT_ID` | - | OAuth client ID (for `jwt` mode) |
| `TOM_OAUTH_CLIENT_SECRET` | - | OAuth client secret (required for Google only) |
| `TOM_OAUTH_DISCOVERY_URL` | - | OIDC discovery URL (for `jwt` mode) |
| `TOM_OAUTH_REDIRECT_PORT` | `8899` | OAuth callback port |
| `TOM_OAUTH_SCOPES` | `openid email profile` | OAuth scopes |
| `TOM_CONFIG_DIR` | `~/.tom` | Config directory path |
| `TOM_CACHE_TTL` | `1h` | Inventory cache TTL (duration: 30m, 2h, etc.) |

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

OAuth 2.0 with PKCE. Supports multiple providers:
- **Standard OIDC** (Duo, Okta, Keycloak, Auth0) - No client secret required
- **Google** - Requires client secret
- **Microsoft** (Entra ID / Azure AD) - No client secret required (public client)

**config.json (standard OIDC):**
```json
{
  "auth_mode": "jwt",
  "oauth_provider": "oidc",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://provider/.well-known/openid-configuration"
}
```

**config.json (Google):**
```json
{
  "auth_mode": "jwt",
  "oauth_provider": "google",
  "oauth_client_id": "xxx.apps.googleusercontent.com",
  "oauth_client_secret": "GOCSPX-xxxxx",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration"
}
```

**config.json (Microsoft):**
```json
{
  "auth_mode": "jwt",
  "oauth_provider": "microsoft",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://login.microsoftonline.com/your-tenant-id/v2.0/.well-known/openid-configuration"
}
```

```bash
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
  "auth_mode": "jwt",
  "oauth_provider": "oidc",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://your-provider/.well-known/openid-configuration"
}
EOF

# Authenticate
tomclient auth login

# Use it
tomclient device router1 "show version"
```

### OAuth Authentication

**Using config.json (recommended):**
```bash
# Create config
mkdir -p ~/.tom
cat > ~/.tom/config.json << EOF
{
  "auth_mode": "jwt",
  "oauth_provider": "oidc",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://your-provider/.well-known/openid-configuration"
}
EOF

# Login
tomclient auth login

# Check status
tomclient auth status

# Use authenticated commands
tomclient device router1 "show version"
tomclient export --filter=switches
```

**Or using environment variables:**
```bash
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_PROVIDER=oidc
export TOM_OAUTH_CLIENT_ID=your-client-id
export TOM_OAUTH_DISCOVERY_URL=https://your-provider/.well-known/openid-configuration

tomclient auth login
tomclient device router1 "show version"
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

## Shell Autocomplete

tomclient provides device name autocomplete for SSH and other commands using cached inventory.

### Quick Setup

**Bash** - Add to `~/.bashrc`:
```bash
eval "$(tomclient completion bash)"
source /path/to/tomclient/shell/ssh_complete.sh
```

**Zsh** - Add to `~/.zshrc`:
```zsh
eval "$(tomclient completion zsh)"
source /path/to/tomclient/shell/ssh_complete.zsh
```

### Usage

```bash
# Populate cache
tomclient inventory --refresh

# SSH autocomplete now works
ssh router<TAB>
# Completes to: router1, router2, router-3, etc.
```

See [AUTOCOMPLETE.md](AUTOCOMPLETE.md) for complete setup guide, troubleshooting, and advanced usage.

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
