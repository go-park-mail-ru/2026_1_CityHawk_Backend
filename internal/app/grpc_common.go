package app

import (
	"context"

	appconfig "cityhawk/backend/internal/config"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newGRPCPostgresPool(cfg appconfig.Config) (*pgxpool.Pool, error) {
	ctx := context.Background()
	pool, err := platformpostgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
