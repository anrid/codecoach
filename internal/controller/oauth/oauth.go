package oauth

import (
	"net/http"

	"github.com/anrid/codecoach/internal/domain"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Controller ...
type Controller struct {
	oa domain.OAuthUseCase
}

// New ...
func New(oa domain.OAuthUseCase) *Controller {
	return &Controller{oa}
}

// SetupRoutes ...
func (co *Controller) SetupRoutes(s *httpserver.HTTPServer) {
	s.Echo.GET("/api/v1/oauth/login-url", co.GetOAuthLoginURL)
	s.Echo.POST("/api/v1/oauth/signup-url", co.PostOAuthSignupURL)
	s.Echo.GET("/api/v1/oauth/callback", co.GetOAuthCallback)
}

// GetOAuthLoginURL ...
func (co *Controller) GetOAuthLoginURL(c echo.Context) error {
	return httpserver.UnescapedJSON(c, http.StatusOK, GetOAuthURLResponse{co.oa.OAuthLoginURL()})
}

// PostOAuthSignupURL ...
func (co *Controller) PostOAuthSignupURL(c echo.Context) (err error) {
	r := new(PostGetOAuthURLRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if err = c.Validate(r); err != nil {
		return httpserver.NewError(http.StatusBadRequest, err, httpserver.GetValidatorError(err))
	}

	url := co.oa.OAuthSignupURL(r.AccountName, r.GivenName, r.FamilyName)

	return httpserver.UnescapedJSON(c, http.StatusOK, GetOAuthURLResponse{url})
}

// PostGetOAuthURLRequest ...
type PostGetOAuthURLRequest struct {
	AccountName string `json:"account_name" validate:"required,gte=2"`
	GivenName   string `json:"given_name" validate:"required,gte=1"`
	FamilyName  string `json:"family_name" validate:"required,gte=1"`
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

	up, err := co.oa.ExchangeCodeForUserProfile(code, state)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not get user profile")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"profile": up,
	})
}

// GetOAuthURLResponse ...
type GetOAuthURLResponse struct {
	URL string `json:"url"`
}
