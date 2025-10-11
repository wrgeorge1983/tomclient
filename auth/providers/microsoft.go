package providers

import (
	"net/url"
)

type MicrosoftProvider struct{}

func (p *MicrosoftProvider) Name() string {
	return "microsoft"
}

func (p *MicrosoftProvider) RequiresClientSecret() bool {
	return false
}

func (p *MicrosoftProvider) BuildTokenRequest(code, verifier, clientID, clientSecret, redirectURI string) url.Values {
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

func (p *MicrosoftProvider) BuildRefreshRequest(refreshToken, clientID, clientSecret string) url.Values {
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

func (p *MicrosoftProvider) AuthURLParams() url.Values {
	return url.Values{}
}
