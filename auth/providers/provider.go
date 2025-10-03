package providers

import (
	"fmt"
	"net/url"
)

type Provider interface {
	Name() string
	BuildTokenRequest(code, verifier, clientID, clientSecret, redirectURI string) url.Values
	RequiresClientSecret() bool
}

func GetProvider(name string) (Provider, error) {
	switch name {
	case "oidc", "":
		return &OIDCProvider{}, nil
	case "google":
		return &GoogleProvider{}, nil
	default:
		return nil, fmt.Errorf("unknown OAuth provider '%s' - must be one of: oidc, google", name)
	}
}
