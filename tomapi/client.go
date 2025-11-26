package tomapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	BaseURL      string
	AuthProvider AuthProvider
	HTTPClient   *http.Client
}

// NewClient creates a new Tom API client with the given auth provider
func NewClient(baseURL string, authProvider AuthProvider) *Client {
	if authProvider == nil {
		authProvider = &NoAuth{}
	}
	return &Client{
		BaseURL:      baseURL,
		AuthProvider: authProvider,
		HTTPClient:   &http.Client{},
	}
}

// NewClientWithAPIKey creates a client with API key authentication
func NewClientWithAPIKey(baseURL, apiKey string) *Client {
	return NewClient(baseURL, &APIKeyAuth{
		APIHeader: "X-API-Key",
		APIKey:    apiKey,
	})
}

// NewClientWithToken creates a client with bearer token authentication
func NewClientWithToken(baseURL, token string) *Client {
	return NewClient(baseURL, &BearerTokenAuth{
		Token: token,
	})
}

func (c *Client) makeRequest(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if err := c.setAuthHeader(req); err != nil {
		return nil, err
	}

	return c.HTTPClient.Do(req)
}

// makeJSONRequest makes an HTTP request with a JSON body
func (c *Client) makeJSONRequest(method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if err := c.setAuthHeader(req); err != nil {
		return nil, err
	}

	return c.HTTPClient.Do(req)
}

func (c *Client) setAuthHeader(req *http.Request) error {
	if c.AuthProvider == nil {
		return nil
	}
	return c.AuthProvider.AddAuth(req)
}
