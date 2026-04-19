package postgres

import (
	"context"

	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func NewPool(ctx context.Context, cfg config.PostgresConfig) (pool *pgxpool.Pool, err error) {
	defer err2.Handle(&err)

	poolConfig := try.To1(pgxpool.ParseConfig(cfg.DSN()))
	poolConfig.MaxConns = int32(cfg.MaxConns)

	pool = try.To1(pgxpool.NewWithConfig(ctx, poolConfig))

	// Test connection
	try.To(pool.Ping(ctx))

	return pool, nil
}
