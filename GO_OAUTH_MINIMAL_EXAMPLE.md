# Minimal OAuth/PKCE Implementation for Go

This is the absolute minimum code needed to implement Duo PKCE authentication in Go.

## 1. PKCE Code Generation (30 lines)

```go
package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
)

// GenerateCodeVerifier creates a random 43-character string
func GenerateCodeVerifier() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64URLEncode(b), nil
}

// GenerateCodeChallenge creates SHA256 hash of verifier
func GenerateCodeChallenge(verifier string) string {
    h := sha256.New()
    h.Write([]byte(verifier))
    return base64URLEncode(h.Sum(nil))
}

// base64URLEncode encodes bytes to base64url (no padding)
func base64URLEncode(data []byte) string {
    return base64.RawURLEncoding.EncodeToString(data)
}
```

## 2. Authorization URL (15 lines)

```go
import (
    "fmt"
    "net/url"
)

// BuildAuthURL creates the Duo authorization URL
func BuildAuthURL(clientID, duoBaseURL, codeChallenge string) string {
    params := url.Values{
        "response_type":         {"code"},
        "client_id":             {clientID},
        "redirect_uri":          {"http://localhost:8899/callback"},
        "code_challenge":        {codeChallenge},
        "code_challenge_method": {"S256"},
        "scope":                 {"openid email profile"},
        "state":                 {generateRandomState()}, // Optional but recommended
    }
    return fmt.Sprintf("%s/oauth/v2/authorize?%s", duoBaseURL, params.Encode())
}
```

## 3. Local Callback Server (50 lines)

```go
import (
    "context"
    "fmt"
    "net/http"
    "time"
)

// StartCallbackServer starts HTTP server and waits for OAuth callback
func StartCallbackServer(ctx context.Context) (string, error) {
    codeChan := make(chan string, 1)
    errChan := make(chan error, 1)

    srv := &http.Server{Addr: ":8899"}
    
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        if code == "" {
            errChan <- fmt.Errorf("no code in callback")
            fmt.Fprintf(w, "<html><body><h1>Authentication Failed</h1></body></html>")
            return
        }
        
        codeChan <- code
        fmt.Fprintf(w, "<html><body><h1>Success!</h1><p>You can close this window.</p></body></html>")
    })

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            errChan <- err
        }
    }()

    defer srv.Shutdown(context.Background())

    select {
    case code := <-codeChan:
        return code, nil
    case err := <-errChan:
        return "", err
    case <-ctx.Done():
        return "", fmt.Errorf("timeout waiting for callback")
    case <-time.After(2 * time.Minute):
        return "", fmt.Errorf("timeout after 2 minutes")
    }
}
```

## 4. Token Exchange (40 lines)

```go
import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "net/url"
)

type TokenResponse struct {
    AccessToken string `json:"access_token"`
    IDToken     string `json:"id_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
}

// ExchangeCodeForToken exchanges auth code for tokens
func ExchangeCodeForToken(duoBaseURL, clientID, code, verifier string) (*TokenResponse, error) {
    data := url.Values{
        "grant_type":    {"authorization_code"},
        "code":          {code},
        "client_id":     {clientID},
        "redirect_uri":  {"http://localhost:8899/callback"},
        "code_verifier": {verifier},
    }

    tokenURL := fmt.Sprintf("%s/oauth/v2/token", duoBaseURL)
    resp, err := http.PostForm(tokenURL, data)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("token exchange failed: %s", body)
    }

    var tokens TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
        return nil, err
    }

    return &tokens, nil
}
```

## 5. Token Storage (40 lines)

```go
import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

type StoredToken struct {
    AccessToken string    `json:"access_token"`
    IDToken     string    `json:"id_token"`
    ExpiresIn   int       `json:"expires_in"`
    ObtainedAt  time.Time `json:"obtained_at"`
    ExpiresAt   time.Time `json:"expires_at"`
}

// SaveToken stores token to ~/.tom/token.json
func SaveToken(token *TokenResponse) error {
    homeDir, _ := os.UserHomeDir()
    tomDir := filepath.Join(homeDir, ".tom")
    os.MkdirAll(tomDir, 0700)

    stored := StoredToken{
        AccessToken: token.AccessToken,
        IDToken:     token.IDToken,
        ExpiresIn:   token.ExpiresIn,
        ObtainedAt:  time.Now(),
        ExpiresAt:   time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
    }

    data, _ := json.MarshalIndent(stored, "", "  ")
    tokenPath := filepath.Join(tomDir, "token.json")
    return os.WriteFile(tokenPath, data, 0600)
}

// LoadToken reads token from ~/.tom/token.json
func LoadToken() (*StoredToken, error) {
    homeDir, _ := os.UserHomeDir()
    tokenPath := filepath.Join(homeDir, ".tom", "token.json")
    
    data, err := os.ReadFile(tokenPath)
    if err != nil {
        return nil, err
    }

    var token StoredToken
    if err := json.Unmarshal(data, &token); err != nil {
        return nil, err
    }

    return &token, nil
}

// IsValid checks if token is not expired (with 60s buffer)
func (t *StoredToken) IsValid() bool {
    return time.Now().Before(t.ExpiresAt.Add(-60 * time.Second))
}
```

## 6. Main Authentication Flow (30 lines)

```go
import (
    "context"
    "fmt"
    "github.com/pkg/browser"
    "time"
)

// Authenticate performs complete OAuth flow
func Authenticate(clientID, duoBaseURL string) error {
    // 1. Generate PKCE codes
    verifier, _ := GenerateCodeVerifier()
    challenge := GenerateCodeChallenge(verifier)

    // 2. Build auth URL and open browser
    authURL := BuildAuthURL(clientID, duoBaseURL, challenge)
    fmt.Println("Opening browser for authentication...")
    browser.OpenURL(authURL)

    // 3. Start local server and wait for callback
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    
    code, err := StartCallbackServer(ctx)
    if err != nil {
        return fmt.Errorf("failed to get auth code: %w", err)
    }

    // 4. Exchange code for tokens
    tokens, err := ExchangeCodeForToken(duoBaseURL, clientID, code, verifier)
    if err != nil {
        return fmt.Errorf("failed to exchange code: %w", err)
    }

    // 5. Save tokens
    if err := SaveToken(tokens); err != nil {
        return fmt.Errorf("failed to save token: %w", err)
    }

    fmt.Println("✅ Authentication successful!")
    return nil
}
```

## 7. API Request Helper (25 lines)

```go
// GetAuthHeader returns the appropriate auth header
func GetAuthHeader() (string, string, error) {
    token, err := LoadToken()
    if err != nil {
        return "", "", fmt.Errorf("not authenticated: %w", err)
    }

    if !token.IsValid() {
        return "", "", fmt.Errorf("token expired, please re-authenticate")
    }

    // Use access_token if available, otherwise id_token
    tokenValue := token.AccessToken
    if tokenValue == "" {
        tokenValue = token.IDToken
    }

    return "Authorization", fmt.Sprintf("Bearer %s", tokenValue), nil
}

// MakeAuthenticatedRequest makes API call with auth
func MakeAuthenticatedRequest(url string) (*http.Response, error) {
    req, _ := http.NewRequest("GET", url, nil)
    
    headerName, headerValue, err := GetAuthHeader()
    if err != nil {
        return nil, err
    }
    
    req.Header.Set(headerName, headerValue)
    return http.DefaultClient.Do(req)
}
```

## 8. CLI Commands (20 lines)

```go
// CLI usage example
func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: tom-cli [auth|test]")
        return
    }

    clientID := os.Getenv("TOM_DUO_CLIENT_ID")
    duoURL := os.Getenv("TOM_DUO_BASE_URL")
    apiURL := os.Getenv("TOM_API_URL")

    switch os.Args[1] {
    case "auth":
        if err := Authenticate(clientID, duoURL); err != nil {
            fmt.Printf("Authentication failed: %v\n", err)
        }
    case "test":
        resp, err := MakeAuthenticatedRequest(apiURL + "/api/")
        if err != nil {
            fmt.Printf("Request failed: %v\n", err)
        } else {
            fmt.Printf("Success! Status: %d\n", resp.StatusCode)
        }
    }
}
```

## Complete Example (All Together)

Total: ~250 lines of Go code

```
auth/
├── pkce.go          (30 lines)  - PKCE generation
├── url.go           (15 lines)  - Auth URL building
├── server.go        (50 lines)  - Callback server
├── token.go         (40 lines)  - Token exchange
├── storage.go       (40 lines)  - Token storage
├── authenticate.go  (30 lines)  - Main auth flow
├── client.go        (25 lines)  - API helper
└── cli.go           (20 lines)  - CLI commands
```

## Dependencies

```go
require (
    github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
)
```

That's it! Standard library handles everything else.

## Testing Your Implementation

```bash
# Set environment
export TOM_DUO_CLIENT_ID="your-client-id"
export TOM_DUO_BASE_URL="https://your-tenant.duosecurity.com"
export TOM_API_URL="http://localhost:8020"

# Authenticate
go run main.go auth

# Test API call
go run main.go test
```

## Key Differences from Python

1. **Context usage** - Go's context package for timeouts
2. **Error handling** - Explicit error returns vs exceptions
3. **JSON tags** - Struct tags for marshaling
4. **HTTP client** - net/http vs requests library
5. **File permissions** - os.WriteFile mode parameter

## What the Python Code Does (Reference)

The Python implementation does exactly this, just with:
- `requests` library instead of `net/http`
- Exceptions instead of error returns
- `webbrowser` module instead of `pkg/browser`
- `json` module instead of `encoding/json`

Map Python → Go:
- `cli_auth_pkce.py:40-55` → `pkce.go`
- `cli_auth_pkce.py:110-145` → `server.go`
- `cli_auth_pkce.py:147-180` → `token.go`
- `tom_cli_auth.py:84-105` → `storage.go`
- `tom_cli_auth.py:107-146` → `authenticate.go`

The logic is identical, just different syntax!
