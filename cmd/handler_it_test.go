//go:build integration
// +build integration

package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/phuangpheth/assessment/track"
	"github.com/stretchr/testify/assert"
)

const PORT = 2500

func TestSaveExpense(t *testing.T) {
	ec := echo.New()
	go func(e *echo.Echo) {
		db, err := sql.Open("postgres", "postgresql://root:password@db_save/expenses_test?sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}

		svc := track.NewService(db)
		h := handler{expenseSvc: svc}
		e.POST("/expenses", h.SaveExpense)
		e.Start(fmt.Sprintf(":%d", PORT))
	}(ec)
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", PORT), 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}

	body := `{
    "amount":30,
    "title":"add-title",
    "note":"add-note",
    "tags":["add-tags"]
  }`

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/expenses", PORT), strings.NewReader(body))
	assert.NoError(t, err)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	assert.NoError(t, err)

	byt, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	want := `{"id":1,"amount":30,"title":"add-title","note":"add-note","tags":["add-tags"]}`

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, want, strings.TrimSpace(string(byt)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = ec.Shutdown(ctx)
	assert.NoError(t, err)
}
