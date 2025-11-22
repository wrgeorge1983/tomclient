package tomapi

import (
	"fmt"
	"net/http"
)

// AuthProvider adds authentication to requests
type AuthProvider interface {
	AddAuth(req *http.Request) error
}

// NoAuth provider for unauthenticated requests
type NoAuth struct{}

// AddAuth does nothing for NoAuth
func (a *NoAuth) AddAuth(req *http.Request) error {
	return nil
}

// APIKeyAuth provider for API key authentication
type APIKeyAuth struct {
	APIHeader string
	APIKey    string
}

func (a *APIKeyAuth) AddAuth(req *http.Request) error {
	if a.APIKey == "" {
		return fmt.Errorf("API key is not set")
	}
	if a.APIHeader == "" {
		return fmt.Errorf("API Key header is not set")
	}
	req.Header.Set(a.APIHeader, a.APIKey)
	return nil
}

// BearerTokenAuth provider for Bearer token authentication (e.g. OAuth2)
type BearerTokenAuth struct {
	Token string
}

func (a *BearerTokenAuth) AddAuth(req *http.Request) error {
	if a.Token == "" {
		return fmt.Errorf("Bearer token is not set")
	}
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}
