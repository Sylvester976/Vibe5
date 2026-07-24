package auth

import (
	"golang.org/x/oauth2"

	"spotify-insights/internal/config"
)

// Scopes cover everything Phase 1 needs: recently-played + currently-playing
// for the poller, top-read/profile for the initial account link. Endpoints
// like audio-features/recommendations/related-artists are deliberately not
// requested — they're unavailable to Spotify Dev Mode apps created after
// November 2024 (see docs/ARCHITECTURE.md's API-limits note).
var spotifyScopes = []string{
	"user-read-recently-played",
	"user-top-read",
	"user-read-currently-playing",
	"user-read-private",
	"user-read-email",
}

func NewOAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.SpotifyClientID,
		ClientSecret: cfg.SpotifyClientSecret,
		RedirectURL:  cfg.SpotifyRedirectURI,
		Scopes:       spotifyScopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}
}
