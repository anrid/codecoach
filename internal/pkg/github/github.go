package github

import (
	"net/http"
	"net/url"

	"github.com/anrid/codecoach/internal/pkg/httpclient"
)

// GetOAuthURL ...
func GetOAuthURL(clientID, redirectURI, state string) string {
	params := url.Values{}

	params.Add("client_id", clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("scope", "read:user user:email")
	params.Add("state", state)

	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// GetCodeExchangeURL ...
func GetCodeExchangeURL(clientID, clientSecret, code, state string) string {
	params := url.Values{}

	params.Add("client_id", clientID)
	params.Add("client_secret", clientSecret)
	params.Add("code", code)
	params.Add("state", state)

	return "https://github.com/login/oauth/access_token?" + params.Encode()
}

// ExchangeCode ...
func ExchangeCode(url string) (*ExchangeCodeResponse, error) {
	r := new(ExchangeCodeResponse)

	_, err := httpclient.CallWithOptions(httpclient.Options{
		Method:          http.MethodPost,
		URL:             url,
		ResponsePayload: r,
		Accept:          "application/json",
	})

	return r, err
}

// ExchangeCodeResponse ...
type ExchangeCodeResponse struct {
	AccessToken string `json:"access_token"`
	Score       string `json:"scope"`
	TokenType   string `json:"token_type"`
}
