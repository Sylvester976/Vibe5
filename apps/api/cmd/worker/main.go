package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/config"
	"spotify-insights/internal/db"
	"spotify-insights/internal/poller"
	"spotify-insights/internal/spotify"
)

const (
	dbConnectAttempts  = 10
	dbConnectDelay     = 2 * time.Second
	schemaWaitAttempts = 15
	schemaWaitDelay    = 2 * time.Second
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// The worker starts concurrently with the api container's `-migrate`
	// run under docker-compose, so the schema may not exist for the first
	// few seconds — retry rather than crash-loop.
	pool, err := db.ConnectWithRetry(ctx, cfg.DatabaseURL, dbConnectAttempts, dbConnectDelay)
	if err != nil {
		log.Fatal("connect to database: ", err)
	}
	defer pool.Close()

	// A successful connection doesn't mean the api container's `-migrate`
	// run has finished yet — wait for the schema too.
	if err := db.WaitForSchema(ctx, pool, schemaWaitAttempts, schemaWaitDelay); err != nil {
		log.Fatal("wait for schema: ", err)
	}

	oauthCfg := auth.NewOAuthConfig(cfg)
	spClient := spotify.New(nil)
	tokens := spotify.NewTokenManager(pool, oauthCfg, cfg.TokenEncryptionKey)

	p := &poller.Poller{
		Pool:          pool,
		Spotify:       spClient,
		Tokens:        tokens,
		Interval:      cfg.PollInterval,
		RollupHourUTC: cfg.RollupHourUTC,
	}

	log.Printf("worker starting: poll interval %s, rollup hour %d UTC", cfg.PollInterval, cfg.RollupHourUTC)
	p.Run(ctx)
	log.Println("worker shutting down")
}
