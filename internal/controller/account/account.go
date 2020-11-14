package user

import (
	"net/http"

	"github.com/anrid/codecoach/internal/domain"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	"github.com/labstack/echo/v4"
)

// Controller ...
type Controller struct {
	a domain.AccountUseCases
}

// New ...
func New(a domain.AccountUseCases) *Controller {
	return &Controller{a}
}

// SetupRoutes ...
func (co *Controller) SetupRoutes(s *httpserver.HTTPServer) {
	s.Private.PATCH("/api/v1/accounts/:account_id", co.PatchAccount)
}

// PatchAccount ...
// @Summary Update an account.
// @Description Update an account.
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Account ID"
// @Param opts body user.PatchAccountRequest true "Patch Account Request"
// @Success 200 {object} domain.Account
// @Failure 400 {object} httpserver.ErrorResponse
// @Router /accounts/{account_id} [patch]
func (co *Controller) PatchAccount(c echo.Context) (err error) {
	id := domain.ID(c.Param("account_id"))

	r := new(PatchAccountRequest)
	if err = httpserver.BindAndValidate(c, r); err != nil {
		return err
	}

	u, err := co.a.Update(c.Request().Context(), id, domain.UpdateAccountArgs(*r))
	if err != nil {
		return httpserver.NewError(http.StatusInternalServerError, err, "could not update account")
	}

	return c.JSON(http.StatusOK, u)
}

// PatchAccountRequest ...
type PatchAccountRequest struct {
	Name string `json:"name" validate:"omitempty,gte=2"`
	Logo string `json:"logo" validate:"omitempty,gte=1"`
}
