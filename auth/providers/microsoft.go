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
	return url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
}
