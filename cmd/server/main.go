package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/habibg1232191/reviewer-assignment-service/config"
	"github.com/habibg1232191/reviewer-assignment-service/internal/api"
	"github.com/habibg1232191/reviewer-assignment-service/internal/repository/postgres"
	"github.com/habibg1232191/reviewer-assignment-service/internal/usecase"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.MustLoad()
	pgCfg := cfg.Postgres
	slog.Info("config loaded", "method", "main", "config", cfg)

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pgCfg.Host, pgCfg.Port, pgCfg.User, pgCfg.Password, pgCfg.DBName, pgCfg.SslMode,
	)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		slog.Error("error openning postgresql", "method", "main", "connection string", connString, "error", err)
		panic("error openning db")
	}
	if err := db.Ping(); err != nil {
		slog.Error("db ping failed", "error", err)
		panic("db ping failed")
	}
	defer db.Close()

	reviewerRepo := postgres.NewPostgresReviewerRepository(db)
	reviewerService := usecase.NewReviwerService(reviewerRepo)
	apiRouter := api.NewAPIRouter(reviewerService)

	srv := &http.Server{
		Handler:      apiRouter,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen error", "error", err)
		}
	}()

	slog.Info("server started", "method", "main", "address", cfg.HTTPServer.Address)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}
}
