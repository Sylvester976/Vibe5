package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/config"
	"spotify-insights/internal/db"
	"spotify-insights/internal/server"
	"spotify-insights/internal/spotify"
)

func main() {
	migrateFlag := flag.Bool("migrate", false, "run database migrations before starting")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config: ", err)
	}

	ctx := context.Background()

	// Matches docker-compose's `command: ["/app/api", "-migrate"]` combined
	// with `restart: unless-stopped` — migrate then keep serving, not exit.
	if *migrateFlag {
		if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatal("run migrations: ", err)
		}
		log.Println("migrations applied")
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("connect to database: ", err)
	}
	defer pool.Close()

	spClient := spotify.New(nil)
	oauthCfg := auth.NewOAuthConfig(cfg)
	tokens := spotify.NewTokenManager(pool, oauthCfg, cfg.TokenEncryptionKey)

	srv := server.New(cfg, pool, spClient, oauthCfg, tokens)

	log.Println("api listening on :" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, srv))
}
