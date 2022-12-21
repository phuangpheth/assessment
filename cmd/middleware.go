package cmd

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ErrInvalidTokenAuth is returned when token authentication was invalid.
var ErrInvalidTokenAuth = errors.New("missing or invalid token authentication")

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tk := c.Request().Header.Get("Authorization")
		if _, err := time.Parse("January 02, 2006", tk); err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"code":    http.StatusUnauthorized,
				"message": ErrInvalidTokenAuth.Error(),
			})
		}
		return next(c)
	}
}
