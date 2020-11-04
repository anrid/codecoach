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

const githubAPIV3 = "https://api.github.com"
const acceptGithubAPIJSON = "application/vnd.github.v3+json"

// GetAuthenticatedUser ...
func GetAuthenticatedUser(token string) (*UserProfile, error) {
	r := new(UserProfile)

	_, err := httpclient.CallWithOptions(httpclient.Options{
		Method:          http.MethodGet,
		URL:             githubAPIV3 + "/user",
		ResponsePayload: r,
		Accept:          acceptGithubAPIJSON,
		Token:           token,
	})

	return r, err
}

// ExchangeCodeResponse ...
type ExchangeCodeResponse struct {
	AccessToken string `json:"access_token"`
	Score       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// UserProfile ...
type UserProfile struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
