package domain

import (
	"encoding/json"
)

// OAuthUseCase ...
type OAuthUseCase interface {
	OAuthLoginURL() string
	OAuthSignupURL(accountName, givenName, familyName string) string
	ExchangeCodeForUserProfile(code string, state OAuthState) (*ExternalUserProfile, error)
}

// ExternalUserProfile ...
type ExternalUserProfile struct {
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

// OAuthState ...
type OAuthState struct {
	Type        string `json:"type"`
	AccountName string `json:"account_name"`
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	Code        string `json:"code"`
}

// String ...
func (s OAuthState) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

// OAuthStateFromJSONString ...
func OAuthStateFromJSONString(j string) (OAuthState, error) {
	state := OAuthState{}
	err := json.Unmarshal([]byte(j), &state)
	return state, err
}
