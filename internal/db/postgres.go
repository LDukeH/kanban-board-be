package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// connectTimeout bounds dialing the pooler. Without it a cold serverless
	// instance can block on a TCP handshake until the platform kills it.
	connectTimeout = 5 * time.Second

	// Each serverless instance is short-lived and Supabase's pooler enforces a
	// hard client limit, so keep the per-instance footprint deliberately small.
	maxConns = 2
	minConns = 0
)

// NewPostgresPool parses DATABASE_URL into a bounded pgx pool with explicit
// timeouts so a slow or unreachable database surfaces as a request error
// instead of a hung (or platform-killed) function invocation.
func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = time.Minute
	cfg.ConnConfig.ConnectTimeout = connectTimeout
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Ping with a deadline so the handshake cannot exceed connectTimeout. A
	pingCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		fmt.Fprintf(os.Stderr, "database ping failed (continuing): %v\n", err)
	}

	return pool, nil
}
