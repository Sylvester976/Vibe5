// Package config loads environment configuration shared by both the api and
// worker binaries. It is an explicit struct rather than package-level
// globals because two separate binaries load it independently in the same
// process tree (docker compose) and globals would invite subtle divergence.
package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	DatabaseURL         string

	SessionSecret      []byte
	TokenEncryptionKey []byte
	WidgetHMACSecret   []byte

	PollInterval  time.Duration
	RollupHourUTC int
}

// Load reads configuration from the environment, loading a local .env file
// first if present (docker-compose passes real env vars directly, so a
// missing .env there is expected, not an error).
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// No .env file found — fine under docker-compose, which uses env_file/environment directly.
		_ = err
	}

	cfg := &Config{
		Port:                getEnvDefault("PORT", "8080"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:  os.Getenv("SPOTIFY_REDIRECT_URI"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
	}

	for name, val := range map[string]string{
		"SPOTIFY_CLIENT_ID":     cfg.SpotifyClientID,
		"SPOTIFY_CLIENT_SECRET": cfg.SpotifyClientSecret,
		"SPOTIFY_REDIRECT_URI":  cfg.SpotifyRedirectURI,
		"DATABASE_URL":          cfg.DatabaseURL,
	} {
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable %s", name)
		}
	}

	var err error
	if cfg.SessionSecret, err = decodeKey("SESSION_SECRET"); err != nil {
		return nil, err
	}
	if cfg.TokenEncryptionKey, err = decodeKey("TOKEN_ENCRYPTION_KEY"); err != nil {
		return nil, err
	}
	if len(cfg.TokenEncryptionKey) != 32 {
		return nil, fmt.Errorf("TOKEN_ENCRYPTION_KEY must decode to exactly 32 bytes (AES-256), got %d", len(cfg.TokenEncryptionKey))
	}
	if cfg.WidgetHMACSecret, err = decodeKey("WIDGET_HMAC_SECRET"); err != nil {
		return nil, err
	}

	pollMinutes, err := strconv.Atoi(getEnvDefault("POLL_INTERVAL_MINUTES", "20"))
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL_MINUTES: %w", err)
	}
	cfg.PollInterval = time.Duration(pollMinutes) * time.Minute

	rollupHour, err := strconv.Atoi(getEnvDefault("ROLLUP_HOUR_UTC", "3"))
	if err != nil {
		return nil, fmt.Errorf("invalid ROLLUP_HOUR_UTC: %w", err)
	}
	if rollupHour < 0 || rollupHour > 23 {
		return nil, fmt.Errorf("ROLLUP_HOUR_UTC must be 0-23, got %d", rollupHour)
	}
	cfg.RollupHourUTC = rollupHour

	return cfg, nil
}

func decodeKey(envVar string) ([]byte, error) {
	raw := os.Getenv(envVar)
	if raw == "" {
		return nil, fmt.Errorf("missing required environment variable %s", envVar)
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be valid base64: %w", envVar, err)
	}
	return key, nil
}

func getEnvDefault(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}
