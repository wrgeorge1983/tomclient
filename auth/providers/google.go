package providers

import (
	"net/url"
)

type GoogleProvider struct {
	UseRefreshToken bool
}

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

func (p *GoogleProvider) BuildRefreshRequest(refreshToken, clientID, clientSecret string) url.Values {
	return url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
}

func (p *GoogleProvider) AuthURLParams() url.Values {
	vals := url.Values{}
	if p.UseRefreshToken {
		vals.Set("access_type", "offline")
		vals.Set("prompt", "consent")
	}
	return vals
}
