package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	env "github.com/gonzaloccnc/marketplace-go/config"
	"github.com/gonzaloccnc/marketplace-go/internal/product"
	"github.com/gonzaloccnc/marketplace-go/internal/security"
	"github.com/gonzaloccnc/marketplace-go/internal/users"
	"github.com/gonzaloccnc/marketplace-go/pkg/database"
	"github.com/gonzaloccnc/marketplace-go/pkg/httpx"
)

const (
	// startupTimeout bounds the one-shot bootstrap work (connect, ping,
	// migrations) so the process fails fast instead of hanging on an
	// unreachable database and lets the orchestrator restart it.
	startupTimeout = 15 * time.Second

	// shutdownTimeout bounds how long we wait to drain in-flight requests on
	// SIGTERM before forcing the server closed.
	shutdownTimeout = 10 * time.Second
)

func main() {
	PORT := env.GetOrDefault("ADDR", ":8090")

	// Bootstrap with a deadline: a slow or down database must not hang boot.
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), startupTimeout)
	defer cancelStartup()

	marketplaceDB, err := database.NewMarketplaceDB(startupCtx)
	if err != nil {
		slog.Error("failed to create database config", "error", err)
		os.Exit(1)
	}
	defer marketplaceDB.Close()

	if err := marketplaceDB.Ping(startupCtx); err != nil {
		slog.Error("error during ping database", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(startupCtx, marketplaceDB); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Report json field names (e.g. "email") in binding validation errors.
	httpx.UseJSONFieldNames()

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		httpx.WriteSuccess(c, http.StatusOK, gin.H{"message": "server is alive"})
	})

	api := r.Group("/api")
	product.Register(api, marketplaceDB)
	users.Register(api, marketplaceDB)
	security.Register(
		api,
		users.NewCredentialsFinder(marketplaceDB),
		users.NewRegistrarAdapter(marketplaceDB),
	)

	// Cancel this context on SIGINT/SIGTERM so we can shut down gracefully
	// (every deploy and `docker stop` sends SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: PORT, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "error", err)
			stop()
		}
	}()
	slog.Info("server listening", "addr", PORT)

	<-ctx.Done()
	slog.Info("shutdown signal received, draining connections")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
