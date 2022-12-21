package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/phuangpheth/assessment/track"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func failOnError(err error, message string) {
	if err != nil {
		log.Printf("%s: %s", message, err)
		os.Exit(1)
	}
}

func Execute() {
	ctx := context.Background()
	zLog, err := zap.NewProduction()
	failOnError(err, "failed to new zap.NewProduction")
	defer zLog.Sync()

	zap.ReplaceGlobals(zLog)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	failOnError(err, "failed to connect to database")
	defer db.Close()

	_, err = db.ExecContext(ctx, `
	  CREATE TABLE IF NOT EXISTS expenses (
	    id SERIAL PRIMARY KEY,
	    title TEXT,
	    amount FLOAT,
	    note TEXT,
	    tags TEXT[]
	  );
	`)
	failOnError(err, "failed to create table expenses")

	svc := track.NewService(db)
	e := echo.New()

	err = NewHandler(e, svc)
	failOnError(err, "failed to create handler")

	errChan := make(chan error, 1)
	go func() {
		errChan <- e.Start(fmt.Sprintf(":%s", getEnv("PORT", "3001")))
	}()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			zLog.Fatal("failed to start the server", zap.Error(err))
		}
		zLog.Info("shutdown server gracefully")

	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		zLog.Info("shutting down the server")
		if err := e.Shutdown(ctx); err != nil {
			zLog.Fatal("failed to shutdown the server", zap.Error(err))
		}
		zLog.Info("shutdown server gracefully")
	}
}
