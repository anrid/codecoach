package user

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/anrid/codecoach/internal/config"
	"github.com/anrid/codecoach/internal/domain"
	"github.com/anrid/codecoach/internal/pkg/github"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	"github.com/anrid/codecoach/internal/pkg/token"
	token_gen "github.com/anrid/codecoach/internal/pkg/token"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Controller ...
type Controller struct {
	a      domain.AccountDAO
	u      domain.UserDAO
	c      *config.Config
	states map[string]bool
	mux    *sync.Mutex
}

// New ...
func New(a domain.AccountDAO, u domain.UserDAO, c *config.Config) *Controller {
	return &Controller{a, u, c, make(map[string]bool), new(sync.Mutex)}
}

// SetupRoutes ...
func (co *Controller) SetupRoutes(s *httpserver.HTTPServer) {
	s.Echo.POST("/api/v1/signup", co.Signup)
	s.Echo.POST("/api/v1/login", co.Login)
	s.Echo.POST("/api/v1/accounts/:account_id/users", co.PostUser)
	s.Echo.PATCH("/api/v1/accounts/:account_id/users/:id", co.PatchUser)
	s.Echo.GET("/api/v1/accounts/:account_id/secret", co.GetSecret)
	s.Echo.GET("/api/v1/oauth/url", co.GetOAuthURL)
	s.Echo.GET("/api/v1/oauth/callback", co.GetOAuthCallback)
}

// GetOAuthURL ...
func (co *Controller) GetOAuthURL(c echo.Context) error {
	state := co.newState("login")

	url := github.GetOAuthURL(co.c.GithubClientID, co.c.GithubRedirectURI, state.String())

	return httpserver.UnescapedJSON(c, http.StatusOK, GetOAuthURLResponse{url})
}

// GetOAuthCallback ...
func (co *Controller) GetOAuthCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return httpserver.NewError(http.StatusUnauthorized, errors.New("missing code"), "missing code")
	}
	stateStr := c.QueryParam("state")

	_, err := co.checkState(stateStr)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, "missing or incorrect state")
	}

	url := github.GetCodeExchangeURL(co.c.GithubClientID, co.c.GithubClientSecret, code, stateStr)

	r1, err := github.ExchangeCode(url)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, "could not exchange oauth code for access token")
	}

	userProfile, err := github.GetAuthenticatedUser(r1.AccessToken)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not get user profile")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"profile": userProfile,
	})
}

func (co *Controller) newState(typ string) oauthState {
	s := oauthState{
		Type: typ,
		Code: token.NewCode(16),
	}

	co.mux.Lock()
	co.states[s.Code] = true
	co.mux.Unlock()

	return s
}

func (co *Controller) checkState(state string) (oauthState, error) {
	s, err := fromString(state)
	if err != nil {
		return s, errors.Wrap(err, "incorrect state")
	}

	co.mux.Lock()
	if _, found := co.states[s.Code]; !found {
		return s, errors.New("missing state")
	}
	delete(co.states, s.Code)
	co.mux.Unlock()

	return s, nil
}

type oauthState struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

func (s oauthState) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func fromString(str string) (oauthState, error) {
	s := oauthState{}
	err := json.Unmarshal([]byte(str), &s)
	return s, err
}

// GetOAuthURLResponse ...
type GetOAuthURLResponse struct {
	URL string `json:"url"`
}

// GetSecret ...
func (co *Controller) GetSecret(c echo.Context) error {
	cu, err := co.requireSession(c)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(http.StatusOK, GetSecretResponse{
		AccountID: cu.AccountID,
		ID:        cu.ID,
		Secret:    "All your base are belong to us",
	})
}

// GetSecretResponse ...
type GetSecretResponse struct {
	AccountID domain.ID `json:"account_id"`
	ID        domain.ID `json:"id"`
	Secret    string    `json:"secret"`
}

func (co *Controller) requireSession(c echo.Context) (*domain.User, error) {
	accountID := c.Param("account_id")

	auth := c.Request().Header.Get("authorization")

	if auth == "" || !strings.HasPrefix(auth, "Bearer ") || len(auth) <= 10 {
		return nil, errors.New("token invalid")
	}

	token := auth[7:]

	u, err := co.u.GetByToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "token invalid")
	}

	if domain.ID(accountID) != u.AccountID {
		return nil, errors.New("account invalid")
	}

	if u.TokenExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return u, nil
}

// PostUser ...
func (co *Controller) PostUser(c echo.Context) (err error) {
	r := new(PostUserRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if err = c.Validate(r); err != nil {
		return httpserver.NewError(http.StatusBadRequest, err, httpserver.GetValidatorError(err))
	}

	cu, err := co.requireSession(c)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, err.Error())
	}

	u := domain.NewUser(cu.AccountID, "user", "user", r.Email, r.Password, r.Role)

	err = co.u.Create(u)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not create user")
	}

	return c.JSON(http.StatusOK, u)
}

// PostUserRequest ...
type PostUserRequest struct {
	Email    string      `json:"email" validate:"required,email"`
	Password string      `json:"password" validate:"required,gte=8"`
	Role     domain.Role `json:"role" validate:"required,oneof=admin hiring_manager candidate"`
}

// Signup ...
func (co *Controller) Signup(c echo.Context) (err error) {
	r := new(SignupRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if err = c.Validate(r); err != nil {
		return httpserver.NewError(http.StatusBadRequest, err, httpserver.GetValidatorError(err))
	}

	// Create new account.
	a, err := domain.NewAccount(r.AccountName)
	if err != nil {
		return httpserver.NewError(http.StatusBadRequest, err, "could not create account code")
	}

	// Create account admin.
	u := domain.NewUser(a.ID, "admin", "admin", r.Email, r.Password, domain.AdminRole)

	// Add admin to account.
	a.AddMember(domain.Member{
		ID:      u.ID,
		Role:    domain.AdminRole,
		AddedAt: time.Now(),
	})
	a.OwnerID = u.ID

	// Save account.
	err = co.a.Create(a)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not create account")
	}

	// Generate access token.
	u.Token = token_gen.New()

	// Make sure token expires at some point in the future.
	expiresAt := time.Now().Add(co.c.TokenExpires)
	u.TokenExpiresAt = &expiresAt

	err = co.u.Create(u)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not create user")
	}

	// Log successful signup.
	zap.S().Infow(
		"signup successful",
		"account", a.ID,
		"user", u.ID,
		"email", u.Email,
		"token_expires", u.TokenExpiresAt,
	)

	return c.JSON(http.StatusOK, SignupResponse{
		Account: a,
		User:    u,
		Token:   u.Token,
	})
}

// SignupRequest ...
type SignupRequest struct {
	AccountName string `json:"account_name" validate:"required,gte=2"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,gte=8"`
}

// SignupResponse ...
type SignupResponse struct {
	Account *domain.Account `json:"account"`
	User    *domain.User    `json:"user"`
	Token   string          `json:"token"`
}

// Login ...
func (co *Controller) Login(c echo.Context) (err error) {
	r := new(LoginRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if err = c.Validate(r); err != nil {
		return httpserver.NewError(http.StatusBadRequest, err, httpserver.GetValidatorError(err))
	}

	// Get account by account code.
	code := domain.CreateCode(r.AccountCode)
	if len(code) < 2 {
		return httpserver.NewError(http.StatusUnauthorized, errors.Errorf("invalid account code '%s'", code), "invalid account, email or password")
	}

	// Get account by account code.
	a, err := co.a.GetByCode(code)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, "invalid account, email or password")
	}

	// Get user by email and account id.
	u, err := co.u.GetByEmail(a.ID, r.Email)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, "invalid account, email or password")
	}

	// Verify password.
	err = u.CheckPassword(r.Password)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, "invalid account, email or password")
	}

	// Generate access token.
	token := token_gen.New()

	// Make sure token expires at some point in the future.
	tokenExpires := time.Now().Add(co.c.TokenExpires)

	// Update user's token.
	_, err = co.u.Update(u.AccountID, u.ID, []domain.Field{
		{Name: "token", Value: token},
		{Name: "token_expires_at", Value: tokenExpires},
	})
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not create token")
	}

	// Log successful login.
	zap.S().Infow(
		"login successful",
		"account", u.AccountID,
		"user", u.ID,
		"email", u.Email,
		"token_expires", tokenExpires.String(),
	)

	return c.JSON(http.StatusOK, LoginResponse{
		Account: a,
		User:    u,
		Token:   token,
	})
}

// LoginRequest ...
type LoginRequest struct {
	AccountCode string `json:"account_code" validate:"required,gte=2"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,gte=8"`
}

// LoginResponse ...
type LoginResponse struct {
	Account *domain.Account `json:"account"`
	User    *domain.User    `json:"user"`
	Token   string          `json:"token"`
}

// PatchUser ...
func (co *Controller) PatchUser(c echo.Context) (err error) {
	accountID := domain.ID(c.Param("account_id"))
	id := domain.ID(c.Param("id"))

	r := new(PatchUserRequest)
	if err = c.Bind(r); err != nil {
		return
	}
	if err = c.Validate(r); err != nil {
		return httpserver.NewError(http.StatusBadRequest, err, httpserver.GetValidatorError(err))
	}

	cu, err := co.requireSession(c)
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, err.Error())
	}

	// Admins can update any user in their account.
	if cu.Role != domain.AdminRole {
		if cu.ID != id {
			err := errors.Errorf("current user %s.%s (role: %s) cannot update user %s.%s", cu.AccountID, cu.ID, cu.Role, accountID, id)
			return httpserver.NewError(http.StatusUnauthorized, err, "access denied")
		}
	}

	var updates []domain.Field

	if r.Email != "" {
		updates = append(updates, domain.Field{Name: "email", Value: r.Email})
	}
	if r.Password != "" {
		cu.SetPassword(r.Password)
		updates = append(updates, domain.Field{Name: "password_hash", Value: cu.PasswordHash})
	}

	updated, err := co.u.Update(cu.AccountID, id, updates)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not update user")
	}

	return c.JSON(http.StatusOK, updated)
}

// PatchUserRequest ...
type PatchUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,gte=8"`
}
