package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/anrid/codecoach/internal/domain"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// HTTPServer ...
type HTTPServer struct {
	Echo    *echo.Echo
	Private *echo.Group
}

// New ...
func New(up UserProvider) *HTTPServer {
	// Echo instance.
	e := echo.New()

	// Setup custom validator.
	e.Validator = &customValidator{validator.New()}

	// Setup custom error handler.
	e.HTTPErrorHandler = customErrorHandler

	// Middleware.
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	g := e.Group("", NewAuthMiddleware(up))

	return &HTTPServer{e, g}
}

// Start ...
func (s *HTTPServer) Start(host string) {
	s.Echo.Logger.Fatal(s.Echo.Start(host))
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// NewError ...
func NewError(code int, err error, message string) error {
	// Log error.
	zap.S().Infow("server error", "error", err.Error())

	if est, ok := err.(stackTracer); ok {
		// Found stack trace. Dump'em!
		st := est.StackTrace()
		max := 3
		if max > len(st) {
			max = len(st)
		}

		fmt.Printf("=== STACK TRACE ===\n")
		fmt.Printf("%+v\n", st[0:max]) // Print max stack frames.
		fmt.Printf("===================\n")
	}

	return echo.NewHTTPError(code, message)
}

func customErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	var message interface{} = err.Error()

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}

	c.JSON(code, errorResponse{message})
}

// GetValidatorError ...
func GetValidatorError(err error) string {
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range verrs {
			return "validation error: " + strings.ToLower(e.Field())
		}
	}
	return ""
}

type errorResponse struct {
	Error interface{} `json:"error"`
}

type customValidator struct {
	v *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}

// UnescapedJSON returns a JSON payload that doesn't escape HTML.
// Go v1.7 added this:
// encoding/json: add Encoder.DisableHTMLEscaping This provides
// a way to disable the escaping of <, >, and & in JSON strings.
func UnescapedJSON(c echo.Context, code int, payload interface{}) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(code)

	enc := json.NewEncoder(c.Response())
	enc.SetEscapeHTML(false)

	return enc.Encode(payload)
}

// UserProvider ...
type UserProvider interface {
	GetByToken(ctx context.Context, token string) (*domain.User, error)
}

// NewAuthMiddleware creates a new instance of Echo middleware
// that checks for the presence of a token in the Authorization
// HTTP header (e.g. `Authorization: Bearer XXX`) and looks up
// a user by the token.
func NewAuthMiddleware(up UserProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			accountID := c.Param("account_id")
			if accountID == "" || len(accountID) <= 10 {
				return errors.New("account id missing or invalid")
			}

			auth := c.Request().Header.Get("authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") || len(auth) <= 10 {
				return errors.New("token invalid")
			}

			token := auth[7:]

			u, err := up.GetByToken(ctx, token)
			if err != nil {
				return errors.Wrap(err, "token invalid")
			}

			if domain.ID(accountID) != u.AccountID {
				return errors.New("account invalid")
			}

			if u.TokenExpiresAt.Before(time.Now()) {
				return errors.New("token expired")
			}

			// Replace current request object
			c.SetRequest(c.Request().WithContext(domain.ContextWithSession(ctx, u)))

			return next(c)
		}
	}
}
