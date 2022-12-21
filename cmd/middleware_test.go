package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/expenses", nil)

	t.Run("Auth()", func(t *testing.T) {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, time.Now().Format("January 02, 2006"))

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := Auth(func(c echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		err := h(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("Auth() returns unauthorized", func(t *testing.T) {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "January 02, 2006 55555")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := Auth(func(c echo.Context) error {
			return c.String(http.StatusOK, "test")
		})

		want := `{"code":401,"message":"missing or invalid token authentication"}`

		err := h(c)
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, want, strings.TrimSpace(rec.Body.String()))
		}
	})
}
