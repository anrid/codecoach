package httpserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// HTTPServer ...
type HTTPServer struct {
	Echo *echo.Echo
}

// New ...
func New() *HTTPServer {
	// Echo instance.
	e := echo.New()

	// Setup custom validator.
	e.Validator = &customValidator{validator.New()}

	// Setup custom error handler.
	e.HTTPErrorHandler = customErrorHandler

	// Middleware.
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return &HTTPServer{e}
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
