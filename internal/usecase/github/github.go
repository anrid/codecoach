package github

import (
	"sync"

	"github.com/anrid/codecoach/internal/config"
	"github.com/anrid/codecoach/internal/domain"
	"github.com/anrid/codecoach/internal/pkg/github"
	"github.com/anrid/codecoach/internal/pkg/token"
	"github.com/pkg/errors"
)

// UseCase ...
type UseCase struct {
	c      *config.Config
	states map[string]bool
	mux    *sync.Mutex
}

var _ domain.OAuthUseCases = &UseCase{}

// New ...
func New(c *config.Config) *UseCase {
	return &UseCase{
		c,
		make(map[string]bool),
		new(sync.Mutex),
	}
}

// OAuthLoginURL ...
func (u *UseCase) OAuthLoginURL(accountCode string) string {
	state := domain.OAuthState{
		Type:        "login",
		AccountCode: accountCode,
		Code:        token.NewCode(16),
	}

	return github.GetOAuthURL(u.c.GithubClientID, u.c.GithubRedirectURI, state.String())
}

// OAuthSignupURL ...
func (u *UseCase) OAuthSignupURL(accountName, givenName, familyName string) string {
	state := domain.OAuthState{
		Type:        "signup",
		AccountName: accountName,
		GivenName:   givenName,
		FamilyName:  familyName,
		Code:        token.NewCode(16),
	}

	return github.GetOAuthURL(u.c.GithubClientID, u.c.GithubRedirectURI, state.String())
}

// ExchangeCodeForUserProfile ...
func (u *UseCase) ExchangeCodeForUserProfile(code string, state domain.OAuthState) (*domain.ExternalUserProfile, error) {
	url := github.GetCodeExchangeURL(u.c.GithubClientID, u.c.GithubClientSecret, code, state.String())

	res, err := github.ExchangeCode(url)
	if err != nil {
		return nil, errors.Wrap(err, "could not exchange oauth code for access token")
	}

	api := github.New(res.AccessToken)

	gu, err := api.CurrentUser()
	if err != nil {
		return nil, errors.Wrap(err, "could not get user profile")
	}

	// Convert.
	eup := domain.ExternalUserProfile(*gu)

	return &eup, nil
}
