# Authentication Guide

The `tomclient` CLI supports three authentication modes for the Tom API.

## Authentication Modes

### 1. None (Default)
No authentication - for testing or when the API is running in `auth_mode=none`.

```bash
export TOM_AUTH_MODE=none
tomclient device router1 "show version"
```

### 2. API Key
Use a static API key for authentication.

**Required Configuration:**
- `TOM_AUTH_MODE=api_key`
- `TOM_API_KEY=<your-api-key>`

```bash
export TOM_AUTH_MODE=api_key
export TOM_API_KEY=your-secret-key-here
tomclient device router1 "show version"
```

### 3. JWT (OAuth with PKCE)
Use OAuth authentication with any OIDC-compliant provider (Duo, Google, Microsoft, etc.).

**Required Configuration:**
- `TOM_AUTH_MODE=jwt`
- `TOM_OAUTH_CLIENT_ID=<your-client-id>`
- `TOM_OAUTH_DISCOVERY_URL=<your-oidc-discovery-url>`

**Optional Configuration:**
- `TOM_OAUTH_REDIRECT_PORT=8899` (default)
- `TOM_OAUTH_SCOPES="openid email profile"` (default)

The discovery URL is the full path to your provider's OIDC discovery document (`.well-known/openid-configuration`).

**Examples:**
```bash
# Duo
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_CLIENT_ID=your-client-id
export TOM_OAUTH_DISCOVERY_URL=https://sso-xxxxx.sso.duosecurity.com/oidc/your-client-id/.well-known/openid-configuration

# Google
export TOM_OAUTH_DISCOVERY_URL=https://accounts.google.com/.well-known/openid-configuration

# Microsoft
export TOM_OAUTH_DISCOVERY_URL=https://login.microsoftonline.com/your-tenant-id/v2.0/.well-known/openid-configuration

# Okta
export TOM_OAUTH_DISCOVERY_URL=https://your-domain.okta.com/.well-known/openid-configuration

# Authenticate (opens browser)
tomclient auth login

# Check status
tomclient auth status

# Use authenticated commands
tomclient device router1 "show version"

# Logout
tomclient auth logout
```

## Configuration Priority

Configuration is loaded in this order (later sources override earlier):

1. Config file (`~/.tom/config.json`)
2. Environment variables (`TOM_*`)
3. Command-line flags (e.g., `--api-url`)

For example, if you set `api_url` in the config file but also pass `--api-url` on the command line, the command-line flag takes precedence.

## Config File

The CLI stores configuration in `~/.tom/config.json`:

```json
{
  "api_url": "http://localhost:8020",
  "auth_mode": "jwt",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration",
  "oauth_redirect_port": 8899
}
```

All configuration options can be set in the config file, including `api_url`.

**Override config directory:**
```bash
tomclient --config-dir=/path/to/config auth status
```

Or set environment variable:
```bash
export TOM_CONFIG_DIR=/path/to/config
```

## Token Storage

When using JWT auth, tokens are stored in `~/.tom/token.json` with `0600` permissions.

The token file contains:
```json
{
  "access_token": "eyJ...",
  "id_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "obtained_at": "2025-01-20T10:30:00Z",
  "expires_at": "2025-01-20T11:30:00Z"
}
```

Tokens are automatically validated before each request. If expired, you'll see:
```
Error: token expired - run 'tomclient auth login' to re-authenticate
```

## Development/Testing

For development, use a local config directory:

```bash
# Create test config
mkdir -p .tom-test
cat > .tom-test/config.json << EOF
{
  "auth_mode": "jwt",
  "duo_client_id": "test-client-id",
  "duo_base_url": "https://api-test.duosecurity.com"
}
EOF

# Use it
tomclient --config-dir=.tom-test auth status
```

## Commands

### `auth login`
Authenticate with OAuth (requires `auth_mode=jwt`).

Opens browser for authentication, saves token locally.

```bash
tomclient auth login
```

### `auth status`
Show current authentication configuration and status.

```bash
tomclient auth status
```

Example output:
```
Auth Mode: jwt
Config Dir: /home/user/.tom
Duo Client ID: your-client-id
Duo Base URL: https://api-xxxxx.duosecurity.com
Status: ✅ Authenticated (expires in 55m30s)
Token Type: Bearer
```

### `auth logout`
Clear stored authentication token.

```bash
tomclient auth logout
```

## Error Messages

### Clear Error Handling

The CLI provides clear, actionable error messages:

**Missing configuration:**
```
Error: auth_mode is 'jwt' but TOM_OAUTH_CLIENT_ID is not set
```

**Not authenticated:**
```
Error: not authenticated - run 'tomclient auth login' first
```

**Token expired:**
```
Error: token expired - run 'tomclient auth login' to re-authenticate
```

### Warnings

Unused configuration triggers warnings:

```bash
TOM_AUTH_MODE=none TOM_API_KEY=unused tomclient device test "show version"
# Output:
# Warning: TOM_API_KEY is set but auth_mode is 'none' - API key will not be used
```

## OAuth Flow Details

The OAuth flow uses PKCE (Proof Key for Code Exchange) with OIDC Discovery:

1. **Discovery**: Fetch the OIDC discovery document from `TOM_OAUTH_DISCOVERY_URL` to get authorization and token endpoints
2. **PKCE Generation**: Generate random `code_verifier` (32 bytes, base64url encoded) and `code_challenge` (SHA256 hash)
3. **Local Server**: Start callback server on `http://localhost:8899/callback`
4. **Authorization**: Open browser to discovered authorization endpoint with challenge
5. **User Authentication**: User authenticates with provider (MFA, SSO, etc.)
6. **Callback**: Browser redirects to localhost with authorization code
7. **Token Exchange**: Exchange code + verifier at discovered token endpoint
8. **Storage**: Save tokens with 60-second expiration buffer

**OIDC Discovery:** You provide the full URL to your provider's `.well-known/openid-configuration` document. The client fetches this once during authentication to discover the `authorization_endpoint` and `token_endpoint`. This makes the client completely provider-agnostic - no special logic for any provider.

**Manual URL fallback:** If browser doesn't open automatically, the URL is printed for manual copy/paste.

## Environment Variables Reference

| Variable | Required For | Default | Description |
|----------|-------------|---------|-------------|
| `TOM_AUTH_MODE` | All modes | `none` | Auth mode: `none`, `api_key`, or `jwt` |
| `TOM_API_KEY` | `api_key` | - | API key for authentication |
| `TOM_OAUTH_CLIENT_ID` | `jwt` | - | OAuth client ID |
| `TOM_OAUTH_DISCOVERY_URL` | `jwt` | - | Full URL to OIDC discovery document |
| `TOM_OAUTH_REDIRECT_PORT` | - | `8899` | Local callback server port |
| `TOM_OAUTH_SCOPES` | - | `openid email profile` | OAuth scopes |
| `TOM_CONFIG_DIR` | - | `~/.tom` | Config directory path |
| `TOM_API_URL` | - | `http://localhost:8000` | Tom API server URL |

## Security Best Practices

1. **Token files have 0600 permissions** - Only your user can read them
2. **Config files have 0600 permissions** - Protects stored credentials
3. **State parameter used** - Prevents CSRF attacks during OAuth flow
4. **60-second expiration buffer** - Tokens refreshed before they expire
5. **Never log tokens** - Sensitive data not printed to console
6. **HTTPS for OAuth** - Token exchange always uses HTTPS (even though redirect is localhost)
