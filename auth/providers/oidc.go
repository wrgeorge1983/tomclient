package providers

import (
	"net/url"
)

type OIDCProvider struct{}

func (p *OIDCProvider) Name() string {
	return "oidc"
}

func (p *OIDCProvider) RequiresClientSecret() bool {
	return false
}

func (p *OIDCProvider) BuildTokenRequest(code, verifier, clientID, clientSecret, redirectURI string) url.Values {
	vals := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
	if clientSecret != "" {
		vals.Set("client_secret", clientSecret)
	}
	return vals
}

func (p *OIDCProvider) BuildRefreshRequest(refreshToken, clientID, clientSecret string) url.Values {
	vals := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}
	if clientSecret != "" {
		vals.Set("client_secret", clientSecret)
	}
	return vals
}

func (p *OIDCProvider) AuthURLParams() url.Values {
	return url.Values{}
}
