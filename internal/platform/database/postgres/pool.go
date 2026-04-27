package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/bbridges_11/document-registry/internal/platform/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// NewPool creates a new PostgreSQL connection pool
// For local profile: uses password authentication
// For cloud profiles: uses IAM authentication with token refresh
func NewPool(ctx context.Context, cfg config.PostgresConfig, profile config.Profile, awsRegion string) (pool *pgxpool.Pool, err error) {
	defer err2.Handle(&err)

	isCloud := cfg.IsCloudProfile(profile)

	if isCloud {
		pool = try.To1(newCloudPool(ctx, cfg, awsRegion))
	} else {
		pool = try.To1(newLocalPool(ctx, cfg))
	}

	// Test connection
	try.To(pool.Ping(ctx))

	return pool, nil
}

// newLocalPool creates a pool for local development (password auth)
func newLocalPool(ctx context.Context, cfg config.PostgresConfig) (pool *pgxpool.Pool, err error) {
	defer err2.Handle(&err)

	// Use existing DSN method (includes password)
	poolConfig := try.To1(pgxpool.ParseConfig(cfg.DSN()))
	poolConfig.MaxConns = int32(cfg.MaxConns)

	pool = try.To1(pgxpool.NewWithConfig(ctx, poolConfig))

	return pool, nil
}

// newCloudPool creates a pool for cloud environments (IAM auth)
func newCloudPool(ctx context.Context, cfg config.PostgresConfig, awsRegion string) (pool *pgxpool.Pool, err error) {
	defer err2.Handle(&err)

	// Build connection string WITHOUT password, SSL required
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=require",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Database,
	)

	poolConfig := try.To1(pgxpool.ParseConfig(connString))
	poolConfig.MaxConns = int32(cfg.MaxConns)

	// Connection lifecycle settings to ensure token refresh
	// MaxConnLifetime forces connections to close before token expires
	poolConfig.MaxConnLifetime = 14 * time.Minute // Close before 15min token expiry
	poolConfig.MaxConnIdleTime = 10 * time.Minute // Close idle connections

	// Create IAM token generator
	// Format: "host:port" for RDS endpoint
	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	tokenGen := newIAMTokenGenerator(endpoint, awsRegion, cfg.User)

	// BeforeConnect hook: called before each new connection
	// This ensures fresh tokens are used for new connections
	poolConfig.BeforeConnect = func(ctx context.Context, connConfig *pgx.ConnConfig) error {
		// Generate/retrieve cached IAM token
		token, err := tokenGen.getToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate IAM auth token: %w", err)
		}

		// Set token as password
		connConfig.Password = token

		return nil
	}

	pool = try.To1(pgxpool.NewWithConfig(ctx, poolConfig))

	return pool, nil
}
