package user

import (
	"net/http"
	"strconv"

	"github.com/anrid/codecoach/internal/domain"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	"github.com/labstack/echo/v4"
)

// Controller ...
type Controller struct {
	u domain.UserUseCases
}

// New ...
func New(u domain.UserUseCases) *Controller {
	return &Controller{u}
}

// SetupRoutes ...
func (co *Controller) SetupRoutes(s *httpserver.HTTPServer) {
	s.Echo.POST("/api/v1/signup", co.Signup)
	s.Echo.POST("/api/v1/login", co.Login)
	s.Private.POST("/api/v1/accounts/:account_id/users", co.PostUser)
	s.Private.GET("/api/v1/accounts/:account_id/users", co.GetList)
	s.Private.PATCH("/api/v1/accounts/:account_id/users/:id", co.PatchUser)
	s.Private.GET("/api/v1/accounts/:account_id/secret", co.GetSecret)
}

// GetSecret ...
// @Summary Get a private test string, used to test user session.
// @Description Get a private test string, used to test user session.
// @Produce json
// @Security Bearer
// @Param account_id path string true "Account ID"
// @Success 200 {object} user.GetSecretResponse
// @Failure 400 {object} httpserver.ErrorResponse
// @Router /accounts/{account_id}/secret [get]
func (co *Controller) GetSecret(c echo.Context) error {
	se, err := domain.RequireSession(c.Request().Context())
	if err != nil {
		return httpserver.NewError(http.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(http.StatusOK, GetSecretResponse{
		AccountID: se.User.AccountID,
		ID:        se.User.ID,
		Secret:    "All your base are belong to us",
	})
}

// GetSecretResponse ...
type GetSecretResponse struct {
	AccountID domain.ID `json:"account_id"`
	ID        domain.ID `json:"id"`
	Secret    string    `json:"secret"`
}

// PostUser ...
// @Summary Create a new user in an account.
// @Description Create a new user in an account.
// @Accept json
// @Produce json
// @Security Bearer
// @Param account_id path string true "Account ID"
// @Param opts body user.PostUserRequest true "Post User Request"
// @Success 200 {object} domain.User
// @Failure 400 {object} httpserver.ErrorResponse
// @Router /accounts/{account_id}/users [post]
func (co *Controller) PostUser(c echo.Context) (err error) {
	r := new(PostUserRequest)
	if err = httpserver.BindAndValidate(c, r); err != nil {
		return err
	}

	u, err := co.u.Create(c.Request().Context(), domain.CreateUserArgs(*r))
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not create user")
	}

	return c.JSON(http.StatusOK, u)
}

// PostUserRequest ...
type PostUserRequest struct {
	GivenName  string      `json:"given_name" validate:"required,gte=1"`
	FamilyName string      `json:"family_name" validate:"required,gte=1"`
	Email      string      `json:"email" validate:"required,email"`
	Password   string      `json:"password" validate:"required,gte=8"`
	Role       domain.Role `json:"role" validate:"required,oneof=admin hiring_manager candidate"`
}

// Signup ...
func (co *Controller) Signup(c echo.Context) (err error) {
	r := new(SignupRequest)
	if err = httpserver.BindAndValidate(c, r); err != nil {
		return err
	}

	// Perform signup.
	res, err := co.u.Signup(c.Request().Context(), domain.SignupArgs{
		AccountName: r.AccountName,
		GivenName:   r.GivenName,
		FamilyName:  r.FamilyName,
		Email:       r.Email,
		Password:    r.Password,
	})
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not perform signup")
	}

	return c.JSON(http.StatusOK, SignupResponse(*res))
}

// SignupRequest ...
type SignupRequest struct {
	AccountName string `json:"account_name" validate:"required,gte=2"`
	GivenName   string `json:"given_name" validate:"required,gte=1"`
	FamilyName  string `json:"family_name" validate:"required,gte=1"`
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
	if err = httpserver.BindAndValidate(c, r); err != nil {
		return err
	}

	res, err := co.u.Login(c.Request().Context(), r.AccountCode, r.Email, r.Password)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not perform login")
	}

	return c.JSON(http.StatusOK, LoginResponse(*res))
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
// @Summary Update a user in an account.
// @Description Update a user in an account.
// @Accept json
// @Produce json
// @Security Bearer
// @Param account_id path string true "Account ID"
// @Param id path string true "User ID"
// @Param opts body user.PatchUserRequest true "Patch User Request"
// @Success 200 {object} domain.User
// @Failure 400 {object} httpserver.ErrorResponse
// @Router /accounts/{account_id}/users/{id} [patch]
func (co *Controller) PatchUser(c echo.Context) (err error) {
	accountID := domain.ID(c.Param("account_id"))
	id := domain.ID(c.Param("id"))

	r := new(PatchUserRequest)
	if err = httpserver.BindAndValidate(c, r); err != nil {
		return err
	}

	u, err := co.u.Update(c.Request().Context(), accountID, id, domain.UpdateUserArgs(*r))
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not update user")
	}

	return c.JSON(http.StatusOK, u)
}

// PatchUserRequest ...
type PatchUserRequest struct {
	GivenName  string `json:"given_name" validate:"omitempty,gte=1"`
	FamilyName string `json:"family_name" validate:"omitempty,gte=1"`
	Email      string `json:"email" validate:"omitempty,email"`
	Password   string `json:"password" validate:"omitempty,gte=8"`
}

// GetList ...
// @Summary Get a list of users in an account.
// @Description Get a list of users in an account.
// @Produce json
// @Security Bearer
// @Param account_id path string true "Account ID"
// @Success 200 {object} user.GetListResponse
// @Failure 400 {object} httpserver.ErrorResponse
// @Router /accounts/{account_id}/users [get]
func (co *Controller) GetList(c echo.Context) error {
	ctx := c.Request().Context()

	pageStr := c.QueryParam("page")
	page := 1
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	res, err := co.u.List(ctx, page)
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not list users")
	}

	return c.JSON(http.StatusOK, GetListResponse{
		Users: res.Users,
		Total: res.Total,
	})
}

// GetListResponse ...
type GetListResponse struct {
	Users      []*domain.User `json:"users"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
	Total      int            `json:"total"`
}
