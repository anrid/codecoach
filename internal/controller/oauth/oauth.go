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
	s.Echo.GET("/api/v1/oauth/url", co.GetOAuthURL)
	s.Echo.GET("/api/v1/oauth/callback", co.GetOAuthCallback)
}

// GetOAuthURL ...
func (co *Controller) GetOAuthURL(c echo.Context) error {
	typ := c.QueryParam("type")
	if typ == "" {
		// Default to login oauth url type.
		typ = "login"
	}

	var url string

	switch typ {
	case "login":
		url = co.oa.OAuthLoginURL()
	case "signup":
		accountName := c.QueryParam("account_name")
		if len(accountName) < 2 {
			return httpserver.NewError(http.StatusBadRequest, errors.Errorf("invalid account name '%s'", accountName), "invalid account name")
		}
		url = co.oa.OAuthSignupURL(accountName)
	default:
		return httpserver.NewError(http.StatusBadRequest, errors.Errorf("invalid oauth url type '%s'", typ), "invalid oauth url type, must be either 'login' or 'signup'")
	}

	return httpserver.UnescapedJSON(c, http.StatusOK, GetOAuthURLResponse{url})
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
