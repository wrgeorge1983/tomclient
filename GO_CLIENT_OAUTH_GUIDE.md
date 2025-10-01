# OAuth/PKCE Integration Guide for Go Clients

This guide shows how to implement PKCE OAuth authentication for a Go CLI client that talks to Tom's API.

## Overview

Tom supports OAuth2 with PKCE (Proof Key for Code Exchange) for secure CLI authentication without client secrets. This has been tested and validated with **Duo Security**.

## What Your Go Client Needs to Do

### 1. PKCE Flow Implementation

```go
// Core PKCE functions you'll need
func GenerateCodeVerifier() string
func GenerateCodeChallenge(verifier string) string
func StartLocalServer() (*http.Server, string, error)  // Returns server, auth code, error
func ExchangeCodeForToken(code, verifier string) (*TokenResponse, error)
```

### 2. Authentication Flow

```
1. Generate PKCE code_verifier (random 43-128 char string)
2. Generate code_challenge = base64url(sha256(code_verifier))
3. Start local HTTP server on localhost:8899/callback
4. Open browser to authorization URL with:
   - response_type=code
   - client_id={your_duo_client_id}
   - redirect_uri=http://localhost:8899/callback
   - code_challenge={generated_challenge}
   - code_challenge_method=S256
   - scope=openid email profile
5. User authenticates in browser (Duo MFA, etc.)
6. Browser redirects to localhost:8899/callback?code={auth_code}
7. Exchange auth_code + code_verifier for tokens
8. Save tokens locally (~/.tom/token.json)
```

### 3. Token Storage

```json
{
  "access_token": "eyJ...",
  "id_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "obtained_at": 1234567890,
  "expires_at": 1234571490
}
```

Store with 0600 permissions at `~/.tom/token.json`

### 4. API Requests

```go
// Load valid token
token, err := LoadValidToken()
if err != nil || token.IsExpired() {
    // Re-authenticate
    token, err = Authenticate()
}

// Make API request
req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
```

## Duo-Specific Details (Tested)

### Authorization URL
```
https://{duo_tenant}.duosecurity.com/oauth/v2/authorize?
  response_type=code&
  client_id={client_id}&
  redirect_uri=http://localhost:8899/callback&
  code_challenge={challenge}&
  code_challenge_method=S256&
  scope=openid+email+profile&
  state={random_state}
```

### Token Exchange URL
```
POST https://{duo_tenant}.duosecurity.com/oauth/v2/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code={auth_code}&
client_id={client_id}&
redirect_uri=http://localhost:8899/callback&
code_verifier={code_verifier}
```

### Token Response
```json
{
  "access_token": "eyJhbG...",
  "id_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "openid email profile"
}
```

### Important Notes for Duo
- **No client secret required** - PKCE only authentication must be enabled in Duo app settings
- **Both tokens work** - Tom accepts either `access_token` or `id_token` in Bearer header
- **Redirect URI** must be exactly `http://localhost:8899/callback` (configure in Duo app)
- **Port** can be configurable but must match what's configured in Duo
- **State parameter** recommended for CSRF protection

## Configuration

### Environment Variables (Recommended)
```bash
TOM_API_URL=http://localhost:8020
TOM_AUTH_MODE=jwt                    # or "hybrid" to support API keys too
TOM_DUO_CLIENT_ID=your-client-id
TOM_DUO_BASE_URL=https://your-tenant.duosecurity.com
TOM_OAUTH_REDIRECT_PORT=8899
```

### Config File (~/.tom/config.json)
```json
{
  "auth_mode": "jwt",
  "api_url": "http://localhost:8020",
  "duo_client_id": "your-client-id",
  "duo_url": "https://your-tenant.duosecurity.com"
}
```

## Error Handling

### Common Errors
1. **Token expired** - Check `expires_at` before each request, re-auth if needed
2. **Invalid token** - Server returns 401, need to re-authenticate
3. **Port in use** - localhost:8899 already occupied, use different port
4. **Browser doesn't open** - Print URL for manual copy/paste
5. **Timeout waiting for callback** - User didn't complete auth in time

### Server Responses
```
200 OK - Authentication successful
401 Unauthorized - Invalid/expired token or missing Authorization header
403 Forbidden - Valid token but insufficient permissions (future RBAC)
```

## Python Reference Implementation

See these files for reference:
- `cli_auth_pkce.py` - Complete PKCE flow implementation
- `tom_cli_auth.py` - CLI integration with config/token management
- `CLI_AUTH_GUIDE.md` - Detailed Python usage guide

### Key Python Code to Reference

**PKCE Generation:**
```python
# From cli_auth_pkce.py:40-55
code_verifier = base64.urlsafe_b64encode(os.urandom(32)).decode('utf-8').rstrip('=')
code_challenge = base64.urlsafe_b64encode(
    hashlib.sha256(code_verifier.encode('utf-8')).digest()
).decode('utf-8').rstrip('=')
```

**Local Server:**
```python
# From cli_auth_pkce.py:110-145
# Simple HTTP server that waits for callback on /callback
# Extracts 'code' parameter from query string
# Returns HTML success/error page to browser
```

**Token Exchange:**
```python
# From cli_auth_pkce.py:147-180
data = {
    'grant_type': 'authorization_code',
    'code': auth_code,
    'client_id': self.client_id,
    'redirect_uri': self.redirect_uri,
    'code_verifier': code_verifier,
}
response = requests.post(token_url, data=data)
```

## Testing Your Implementation

### 1. Test Authentication
```bash
# Should open browser and complete OAuth flow
tom-cli auth login

# Check stored token
cat ~/.tom/token.json
```

### 2. Test API Call
```bash
# Should use stored token
tom-cli --debug api call

# Should see header:
# Authorization: Bearer eyJ...
```

### 3. Test Token Expiration
```bash
# Manually edit token.json to set expires_at to past time
# Next API call should trigger re-authentication
tom-cli api call
```

## Go Libraries to Consider

- **golang.org/x/oauth2** - Standard OAuth2 library, but you'll implement PKCE manually
- **github.com/pkg/browser** - Cross-platform browser opening
- **github.com/dgrijalva/jwt-go** - JWT parsing (optional, for debugging tokens)

## Hybrid Mode Support

If supporting both API keys and JWT:

```go
func GetAuthHeader() (string, string) {
    // Try JWT first
    if token := LoadValidToken(); token != nil && !token.IsExpired() {
        return "Authorization", fmt.Sprintf("Bearer %s", token.AccessToken)
    }
    
    // Fall back to API key
    if apiKey := GetConfiguredAPIKey(); apiKey != "" {
        return "X-API-Key", apiKey
    }
    
    return "", ""
}
```

Tom's server will try JWT first if `Authorization: Bearer` header is present, otherwise fall back to `X-API-Key` in hybrid mode.

## Security Best Practices

1. **Store tokens with 0600 permissions** - Protect token file from other users
2. **Use state parameter** - Prevent CSRF attacks during OAuth flow
3. **Validate redirect** - Ensure callback came from your localhost server
4. **Clear sensitive data** - Don't log tokens or code_verifier
5. **Handle token refresh** - Check expiration before each request
6. **Timeout on auth** - Don't wait forever for user to complete flow
7. **HTTPS in production** - Use HTTPS for token_url, even though redirect_uri is http://localhost

## Future Enhancements (Not Yet Implemented)

- **Token refresh** - Use refresh_token to get new access_token without re-auth
- **RBAC** - Role-based access control from JWT claims
- **Multiple providers** - Google, GitHub, Entra ID (speculative implementations exist but untested)

## Questions?

The Python implementation in this repo is the reference. Key files:
- `cli_auth_pkce.py` - 200 lines, complete PKCE flow
- `tom_cli_auth.py` - 318 lines, config + token management
- `CLI_AUTH_GUIDE.md` - User-facing documentation

Server-side JWT validation is in:
- `services/controller/src/tom_controller/auth/jwt_validator.py` - Base validation
- `services/controller/src/tom_controller/auth/providers.py` - Provider-specific logic
