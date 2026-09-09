package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Abbas-Shahneh/taskManager/internal/config"
)

func NewPostgresPool(
	ctx context.Context,
	cfg config.DatabaseConfig,
) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.PathEscape(cfg.User),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		url.PathEscape(cfg.Name),
		url.QueryEscape(cfg.SSLMode),
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf(
			"parse postgres config: %w",
			err,
		)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime =
		time.Duration(cfg.MaxConnLifetime) * time.Minute
	poolConfig.MaxConnIdleTime =
		time.Duration(cfg.MaxConnIdleTime) * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf(
			"create postgres pool: %w",
			err,
		)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf(
			"ping postgres: %w",
			err,
		)
	}

	return pool, nil
}
