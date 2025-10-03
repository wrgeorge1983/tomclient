# Authentication Guide

The Tom API client supports three authentication modes.

## Authentication Modes

### 1. None (Default)
No authentication headers sent. Use when the Tom API is running with `TOM_CORE_AUTH_MODE=none`.

**config.json:**
```json
{
  "auth_mode": "none"
}
```

**Or via environment:**
```bash
export TOM_AUTH_MODE=none
tomclient device router1 "show version"
```

### 2. API Key
Static API key sent in `X-API-Key` header.

**config.json:**
```json
{
  "auth_mode": "api_key",
  "api_key": "your-secret-key"
}
```

**Or via environment:**
```bash
export TOM_AUTH_MODE=api_key
export TOM_API_KEY=your-secret-key
tomclient device router1 "show version"
```

### 3. OAuth (JWT)
OAuth 2.0 with PKCE for secure authentication. Supports multiple providers with different requirements.

**Standard OIDC Provider (config.json):**
```json
{
  "auth_mode": "jwt",
  "oauth_provider": "oidc",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://your-provider/.well-known/openid-configuration"
}
```

**Google Provider (config.json):**
```json
{
  "auth_mode": "jwt",
  "oauth_provider": "google",
  "oauth_client_id": "xxx.apps.googleusercontent.com",
  "oauth_client_secret": "GOCSPX-xxxxx",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration"
}
```

**Or via environment:**
```bash
export TOM_AUTH_MODE=jwt
export TOM_OAUTH_PROVIDER=google
export TOM_OAUTH_CLIENT_ID=xxx.apps.googleusercontent.com
export TOM_OAUTH_CLIENT_SECRET=GOCSPX-xxxxx
export TOM_OAUTH_DISCOVERY_URL=https://accounts.google.com/.well-known/openid-configuration

# Authenticate (opens browser)
tomclient auth login

# Check authentication status
tomclient auth status

# Run commands (uses stored token)
tomclient device router1 "show version"

# Logout (clear token)
tomclient auth logout
```

## Configuration

### Config File (Recommended)

Store settings in `~/.tom/config.json`:

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

Override location with `--config-dir` flag or `TOM_CONFIG_DIR` environment variable.

### Configuration Precedence

1. **Config file** (`~/.tom/config.json`) - Primary configuration source
2. **Environment variables** (`TOM_*`) - Override config file values
3. **Command-line flags** - Override both config file and environment

### Environment Variables

Environment variables can override config file settings:

| Variable | Required For | Default | Description |
|----------|-------------|---------|-------------|
| `TOM_API_URL` | - | `http://localhost:8000` | Tom API server URL |
| `TOM_AUTH_MODE` | - | `none` | Auth mode: `none`, `api_key`, or `jwt` |
| `TOM_API_KEY` | `api_key` | - | API key for authentication |
| `TOM_OAUTH_PROVIDER` | `jwt` | `oidc` | OAuth provider: `oidc` or `google` |
| `TOM_OAUTH_CLIENT_ID` | `jwt` | - | OAuth client ID |
| `TOM_OAUTH_CLIENT_SECRET` | `jwt` (Google only) | - | OAuth client secret (required for Google) |
| `TOM_OAUTH_DISCOVERY_URL` | `jwt` | - | Full URL to OIDC discovery document |
| `TOM_OAUTH_REDIRECT_PORT` | - | `8899` | Local OAuth callback server port |
| `TOM_OAUTH_SCOPES` | - | `openid email profile` | OAuth scopes |
| `TOM_CONFIG_DIR` | - | `~/.tom` | Config directory path |

## OAuth Details

### How It Works

1. Client fetches OIDC discovery document from `oauth_discovery_url`
2. Browser opens to provider's authorization endpoint
3. User authenticates (MFA, SSO, etc.)
4. Browser redirects to `http://localhost:8899/callback` with authorization code
5. Client exchanges code for tokens using PKCE
6. Token stored in `~/.tom/token.json` with 0600 permissions
7. Token used automatically for subsequent API requests

### OAuth Providers

#### Standard OIDC Provider (oidc)

The default provider uses standard OAuth 2.0 with PKCE (no client secret required). This works with most OIDC-compliant providers:

- **Duo** - `https://sso-xxxxx.sso.duosecurity.com/oidc/client-id/.well-known/openid-configuration`
- **Okta** - `https://your-domain.okta.com/.well-known/openid-configuration`
- **Keycloak** - `https://your-keycloak/realms/your-realm/.well-known/openid-configuration`
- **Auth0** - `https://your-tenant.auth0.com/.well-known/openid-configuration`

**Example config.json:**
```json
{
  "auth_mode": "jwt",
  "oauth_provider": "oidc",
  "oauth_client_id": "your-client-id",
  "oauth_discovery_url": "https://your-provider/.well-known/openid-configuration"
}
```

#### Google Provider (google)

Google requires a client secret even when using PKCE. Create OAuth credentials in Google Cloud Console (select "Web application" type).

**Example config.json:**
```json
{
  "auth_mode": "jwt",
  "oauth_provider": "google",
  "oauth_client_id": "xxx.apps.googleusercontent.com",
  "oauth_client_secret": "GOCSPX-xxxxx",
  "oauth_discovery_url": "https://accounts.google.com/.well-known/openid-configuration"
}
```

**Note:** Set authorized redirect URI to `http://localhost:8899/callback` in Google Cloud Console.

### Token Management

- Tokens expire based on provider settings (typically 1 hour)
- Tokens validated automatically before each request
- 60-second expiration buffer for safety
- Clear error message if token expired: re-run `tomclient auth login`

## Error Messages

| Error | Solution |
|-------|----------|
| `auth_mode is 'jwt' but TOM_OAUTH_CLIENT_ID is not set` | Add `oauth_client_id` to config.json or set `TOM_OAUTH_CLIENT_ID` |
| `OAuth provider 'google' requires client_secret but TOM_OAUTH_CLIENT_SECRET is not set` | Add `oauth_client_secret` to config.json or set `TOM_OAUTH_CLIENT_SECRET` |
| `unknown OAuth provider 'xxx'` | Use `oidc` or `google` as `oauth_provider` |
| `failed to fetch OIDC discovery from ...` | Verify discovery URL is correct and accessible |
| `not authenticated - run 'tomclient auth login' first` | Run `tomclient auth login` |
| `token expired - run 'tomclient auth login'` | Re-authenticate with `tomclient auth login` |

## Security

- Token files stored with 0600 permissions (user-only read/write)
- Config files stored with 0600 permissions
- PKCE flow prevents authorization code interception
- State parameter prevents CSRF attacks
- Tokens never logged to console
