package oauth

import (
	"net/http"
	"strings"

	"github.com/anrid/codecoach/internal/domain"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	"github.com/anrid/codecoach/internal/pkg/token"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Controller ...
type Controller struct {
	o domain.OAuthUseCases
	u domain.UserUseCases
}

// New ...
func New(o domain.OAuthUseCases, u domain.UserUseCases) *Controller {
	return &Controller{o, u}
}

// SetupRoutes ...
func (co *Controller) SetupRoutes(s *httpserver.HTTPServer) {
	s.Echo.POST("/api/v1/oauth/login-url", co.PostOAuthLoginURL)
	s.Echo.POST("/api/v1/oauth/signup-url", co.PostOAuthSignupURL)
	s.Echo.GET("/api/v1/oauth/callback", co.GetOAuthCallback)
}

// PostOAuthLoginURL ...
func (co *Controller) PostOAuthLoginURL(c echo.Context) (err error) {
	r := new(PostOAuthLoginURLRequest)
	if err = httpserver.BindAndValidate(c, r); err != nil {
		return err
	}

	return httpserver.UnescapedJSON(c, http.StatusOK, PostOAuthLoginURLResponse{co.o.OAuthLoginURL(r.AccountCode)})
}

// PostOAuthLoginURLRequest ...
type PostOAuthLoginURLRequest struct {
	AccountCode string `json:"account_code" validate:"required,gte=2"`
}

// PostOAuthLoginURLResponse ...
type PostOAuthLoginURLResponse struct {
	URL string `json:"url"`
}

// PostOAuthSignupURL ...
func (co *Controller) PostOAuthSignupURL(c echo.Context) (err error) {
	r := new(PostOAuthSignupURLRequest)
	if err = httpserver.BindAndValidate(c, r); err != nil {
		return err
	}

	url := co.o.OAuthSignupURL(r.AccountName, r.GivenName, r.FamilyName)

	return httpserver.UnescapedJSON(c, http.StatusOK, PostOAuthSignupURLResponse{url})
}

// PostOAuthSignupURLRequest ...
type PostOAuthSignupURLRequest struct {
	AccountName string `json:"account_name" validate:"required,gte=2"`
	GivenName   string `json:"given_name" validate:"required,gte=1"`
	FamilyName  string `json:"family_name" validate:"required,gte=1"`
}

// PostOAuthSignupURLResponse ...
type PostOAuthSignupURLResponse struct {
	URL string `json:"url"`
}

// GetOAuthCallback ...
func (co *Controller) GetOAuthCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return httpserver.NewError(http.StatusUnauthorized, errors.New("missing code"), "missing code")
	}

	stateStr := c.QueryParam("state")
	if stateStr == "" {
		return httpserver.NewError(http.StatusBadRequest, errors.New("missing oauth state"), "missing oauth state")
	}

	state, err := domain.OAuthStateFromJSONString(stateStr)
	if err != nil {
		return httpserver.NewError(http.StatusBadRequest, errors.Wrap(err, "could not parse oauth state"), "invalid oauth state")
	}

	up, err := co.o.ExchangeCodeForUserProfile(code, state)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not get user profile")
	}

	ctx := c.Request().Context()
	resp := GetOAuthCallbackResponse{}

	switch state.Type {
	case "login":
		// Do the login oauth flow.
		if state.AccountCode != "" {
			// Perform full login if we're passed an account code.
			res, err := co.u.GithubLogin(ctx, state.AccountCode, up.ID)
			if err != nil {
				return httpserver.NewError(http.StatusInternalServerError, err, "could not perform login")
			}

			resp.Type = loginSuccessful
			resp.Account = res.Account
			resp.User = res.User
			resp.Token = res.Token
		} else {
			// Otherwise return all available accounts for user.
			as, err := co.u.GithubGetAvailableAccounts(ctx, up.ID)
			if err != nil {
				return httpserver.NewError(http.StatusInternalServerError, err, "could get available accounts")
			}

			resp.Type = listingAvailableAccounts
			resp.AvailableAccounts = as
		}
		return c.JSON(http.StatusOK, resp)

	case "signup":
		// Do the signup oauth flow.
		givenName := state.GivenName
		familyName := state.FamilyName

		// Use Github profile name if no name is given.
		parts := strings.SplitN(up.Name, " ", -1)
		if givenName == "" && len(parts) > 0 {
			givenName = parts[0]
		}
		if familyName == "" && len(parts) > 0 {
			familyName = parts[len(parts)-1]
		}

		if up.Email == "" {
			// Impossible! All hell breaks loose.
			err = errors.Errorf("no email address found in github profile %#v", up)
			return httpserver.NewError(http.StatusBadRequest, err, "no email address found in github profile")
		}

		// Perform signup.
		res, err := co.u.Signup(ctx, domain.SignupArgs{
			AccountName: state.AccountName,
			Email:       up.Email,
			GivenName:   givenName,
			FamilyName:  familyName,
			Password:    token.NewCode(20), // random unguessable password!
			GithubID:    up.ID,
			GithubLogin: up.Login,
			PhotoURL:    up.AvatarURL,
			Location:    up.Location,
		})
		if err != nil {
			return httpserver.NewError(http.StatusInternalServerError, err, "could not perform signup")
		}

		resp.Type = signupSuccessful
		resp.Account = res.Account
		resp.User = res.User
		resp.Token = res.Token

		return c.JSON(http.StatusOK, resp)

	default:
		return httpserver.NewError(http.StatusBadRequest, err, "invalid oauth state type: "+state.Type)
	}
}

// GetOAuthCallbackResponse ...
type GetOAuthCallbackResponse struct {
	Account           *domain.Account       `json:"account"`
	User              *domain.User          `json:"user"`
	Token             string                `json:"token"`
	AvailableAccounts []*domain.AccountInfo `json:"available_accounts"`
	Type              responseType          `json:"type"`
}

type responseType string

const (
	loginSuccessful          responseType = "login_successful"
	signupSuccessful         responseType = "signup_successful"
	listingAvailableAccounts responseType = "listing_available_accounts"
)
