package providers

import (
	"net/url"
)

type GoogleProvider struct{}

func (p *GoogleProvider) Name() string {
	return "google"
}

func (p *GoogleProvider) RequiresClientSecret() bool {
	return true
}

func (p *GoogleProvider) BuildTokenRequest(code, verifier, clientID, clientSecret, redirectURI string) url.Values {
	return url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
}
