package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"tomclient/auth/providers"
)

type OIDCDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	Issuer                string `json:"issuer"`
}

type OAuthFlow struct {
	Config       *Config
	CodeVerifier string
	State        string
	Discovery    *OIDCDiscovery
	Provider     providers.Provider
}

func discoverOIDCEndpoints(discoveryURL string) (*OIDCDiscovery, error) {
	resp, err := http.Get(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OIDC discovery failed with status %d: %s", resp.StatusCode, string(body))
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to parse OIDC discovery document: %w", err)
	}

	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" {
		return nil, fmt.Errorf("OIDC discovery document missing required endpoints")
	}

	return &discovery, nil
}

func NewOAuthFlow(config *Config) (*OAuthFlow, error) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, err
	}

	state, err := generateState()
	if err != nil {
		return nil, err
	}

	discovery, err := discoverOIDCEndpoints(config.OAuthDiscoveryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC discovery from %s: %w", config.OAuthDiscoveryURL, err)
	}

	provider, err := providers.GetProvider(config.OAuthProvider)
	if err != nil {
		return nil, err
	}

	return &OAuthFlow{
		Config:       config,
		CodeVerifier: verifier,
		State:        state,
		Discovery:    discovery,
		Provider:     provider,
	}, nil
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (f *OAuthFlow) GetAuthURL() string {
	challenge := GenerateCodeChallenge(f.CodeVerifier)

	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {f.Config.OAuthClientID},
		"redirect_uri":          {fmt.Sprintf("http://localhost:%d/callback", f.Config.OAuthRedirectPort)},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"scope":                 {f.Config.OAuthScopes},
		"state":                 {f.State},
	}

	return fmt.Sprintf("%s?%s", f.Discovery.AuthorizationEndpoint, params.Encode())
}

func (f *OAuthFlow) StartCallbackServer(ctx context.Context) (string, error) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	stateChan := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errorParam := r.URL.Query().Get("error")

		if errorParam != "" {
			errChan <- fmt.Errorf("OAuth error: %s", errorParam)
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><h1>Authentication Failed</h1><p>Error: %s</p></body></html>`, errorParam)
			return
		}

		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h1>Authentication Failed</h1><p>No authorization code received</p></body></html>`)
			return
		}

		stateChan <- state
		codeChan <- code
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Authentication Successful!</h1><p>You can close this window and return to your terminal.</p><script>window.setTimeout(function(){window.close()}, 2000);</script></body></html>`)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", f.Config.OAuthRedirectPort),
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	defer srv.Shutdown(context.Background())

	select {
	case code := <-codeChan:
		receivedState := <-stateChan
		if receivedState != f.State {
			return "", fmt.Errorf("state mismatch - possible CSRF attack")
		}
		return code, nil
	case err := <-errChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("authentication timed out")
	case <-time.After(2 * time.Minute):
		return "", fmt.Errorf("authentication timed out after 2 minutes")
	}
}

func (f *OAuthFlow) ExchangeCodeForToken(code string) (*TokenResponse, error) {
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", f.Config.OAuthRedirectPort)
	data := f.Provider.BuildTokenRequest(
		code,
		f.CodeVerifier,
		f.Config.OAuthClientID,
		f.Config.OAuthClientSecret,
		redirectURI,
	)

	resp, err := http.PostForm(f.Discovery.TokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &token, nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

func Authenticate(config *Config) error {
	flow, err := NewOAuthFlow(config)
	if err != nil {
		return fmt.Errorf("failed to initialize OAuth flow: %w", err)
	}

	authURL := flow.GetAuthURL()

	fmt.Println("Opening browser for authentication...")
	fmt.Printf("\nIf the browser doesn't open automatically, visit this URL:\n%s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
		fmt.Println("Please copy and paste the URL above into your browser.")
	}

	fmt.Printf("Waiting for authentication (listening on http://localhost:%d/callback)...\n", config.OAuthRedirectPort)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	code, err := flow.StartCallbackServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive authorization code: %w", err)
	}

	fmt.Println("Authorization code received, exchanging for token...")

	token, err := flow.ExchangeCodeForToken(code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	if err := SaveToken(token, config.ConfigDir, config.OAuthProvider); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("✅ Authentication successful!")
	return nil
}
