package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgx connection pool against databaseURL.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// ConnectWithRetry retries Connect with backoff — covers Postgres itself
// still starting up (docker-compose's service_healthy only guarantees it's
// accepting connections).
func ConnectWithRetry(ctx context.Context, databaseURL string, attempts int, delay time.Duration) (*pgxpool.Pool, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		pool, err := Connect(ctx, databaseURL)
		if err == nil {
			return pool, nil
		}
		lastErr = err
		log.Printf("db connect attempt %d/%d failed: %v", i+1, attempts, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, fmt.Errorf("db connect failed after %d attempts: %w", attempts, lastErr)
}

// WaitForSchema blocks until the users table is queryable or attempts are
// exhausted. The worker binary starts concurrently with the api binary's
// `-migrate` run, so a successful connection doesn't yet mean the schema
// exists — without this, the worker's first poll fails outright and doesn't
// retry until its next scheduled tick (POLL_INTERVAL_MINUTES later).
func WaitForSchema(ctx context.Context, pool *pgxpool.Pool, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		var discard int
		err := pool.QueryRow(ctx, `SELECT 1 FROM users LIMIT 1`).Scan(&discard)
		if err == nil || errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		lastErr = err
		log.Printf("schema not ready yet (attempt %d/%d): %v", i+1, attempts, err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return fmt.Errorf("schema not ready after %d attempts: %w", attempts, lastErr)
}
