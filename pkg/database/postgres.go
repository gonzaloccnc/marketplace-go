package database

import (
	"context"
	"fmt"

	env "github.com/gonzaloccnc/marketplace-go/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewMarketplaceDB(ctx context.Context) (*pgxpool.Pool, error) {
	host := env.GetOrDefault("POSTGRES_HOST", "127.0.0.1")
	port := env.GetOrDefault("POSTGRES_PORT", "5432")
	dbName := env.GetOrDefault("POSTGRES_DB", "mk")
	username := env.GetOrDefault("POSTGRES_USER", "postgres")
	pwd := env.GetOrDefault("POSTGRES_PASSWORD", "postgres")
	sslMode := env.GetOrDefault("POSTGRES_SSL", "disable")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		username, pwd, host, port, dbName, sslMode,
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return pool, nil
}
