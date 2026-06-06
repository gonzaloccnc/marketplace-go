package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	env "github.com/gonzaloccnc/marketplace-go/config"
	"github.com/gonzaloccnc/marketplace-go/internal/product"
	"github.com/gonzaloccnc/marketplace-go/pkg/database"
)

func main() {
	PORT := env.GetOrDefault("ADDR", ":8090")
	marketplaceDB, err := database.NewMarketplaceDB(context.Background())

	if err != nil {
		slog.Error("failed to create database config", "error", err)
		os.Exit(1)
	}

	defer marketplaceDB.Close()

	err = marketplaceDB.Ping(context.Background())

	if err != nil {
		slog.Error("error during ping database", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(context.Background(), marketplaceDB); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "server is alive",
		})
	})

	product.Register(r, marketplaceDB)

	if err := r.Run(PORT); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
