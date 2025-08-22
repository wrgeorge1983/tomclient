package tomapi

import (
	"net/http"
	"os"
)

// Client represents a Tom API client
type Client struct {
	BaseURL   string
	APIKey    string
	KeyHeader string
	HTTPClient *http.Client
}

// NewClient creates a new Tom API client
func NewClient(baseURL string) *Client {
	apiKey := os.Getenv("TOM_API_KEY")
	keyHeader := os.Getenv("TOM_API_KEY_HEADER")
	if keyHeader == "" {
		keyHeader = "X-API-Key"
	}

	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		KeyHeader:  keyHeader,
		HTTPClient: &http.Client{},
	}
}

// NewClientWithAuth creates a new Tom API client with explicit auth
func NewClientWithAuth(baseURL, apiKey, keyHeader string) *Client {
	if keyHeader == "" {
		keyHeader = "X-API-Key"
	}
	
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		KeyHeader:  keyHeader,
		HTTPClient: &http.Client{},
	}
}

// makeRequest creates and executes an HTTP request with authentication
func (c *Client) makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	
	if c.APIKey != "" {
		req.Header.Set(c.KeyHeader, c.APIKey)
	}
	
	return c.HTTPClient.Do(req)
}