// Package poller implements cmd/worker's scheduled collection: pulling each
// connected user's recently-played + currently-playing tracks into
// listening_events, and refreshing cached artist genres.
package poller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"spotify-insights/internal/db"
	"spotify-insights/internal/spotify"
)

const (
	genreStaleAfter  = 30 * 24 * time.Hour
	defaultLookback  = 24 * time.Hour
	rollupCheckEvery = 10 * time.Minute
)

type Poller struct {
	Pool          *pgxpool.Pool
	Spotify       *spotify.Client
	Tokens        *spotify.TokenManager
	Interval      time.Duration
	RollupHourUTC int
}

// Run polls immediately, then on a ticker every Interval, and separately
// checks once per rollupCheckEvery whether it's time for the nightly rollup.
// It blocks until ctx is cancelled.
func (p *Poller) Run(ctx context.Context) {
	p.pollAllUsers(ctx)

	pollTicker := time.NewTicker(p.Interval)
	defer pollTicker.Stop()
	rollupTicker := time.NewTicker(rollupCheckEvery)
	defer rollupTicker.Stop()

	lastRollupDate := ""

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			p.pollAllUsers(ctx)
		case <-rollupTicker.C:
			now := time.Now().UTC()
			today := now.Format("2006-01-02")
			if now.Hour() == p.RollupHourUTC && lastRollupDate != today {
				if err := p.RunNightlyRollup(ctx); err != nil {
					log.Println("nightly rollup:", err)
					continue
				}
				lastRollupDate = today
			}
		}
	}
}

func (p *Poller) pollAllUsers(ctx context.Context) {
	userIDs, err := db.ListUserIDs(ctx, p.Pool)
	if err != nil {
		log.Println("list users:", err)
		return
	}
	for _, userID := range userIDs {
		if err := p.pollUser(ctx, userID); err != nil {
			log.Printf("poll user %s: %v", userID, err)
		}
	}
}

func (p *Poller) pollUser(ctx context.Context, userID uuid.UUID) error {
	accessToken, err := p.Tokens.GetValidAccessToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	since, ok, err := db.LastPlayedAt(ctx, p.Pool, userID)
	if err != nil {
		return fmt.Errorf("last played at: %w", err)
	}
	if !ok {
		since = time.Now().Add(-defaultLookback)
	}

	recentItems, err := p.Spotify.RecentlyPlayed(ctx, accessToken, since)
	if err != nil {
		if rlErr, ok := err.(*spotify.RateLimitError); ok {
			log.Printf("rate limited fetching recently-played for user %s, backing off %s", userID, rlErr.RetryAfter)
			return nil
		}
		return fmt.Errorf("fetch recently played: %w", err)
	}

	current, err := p.Spotify.CurrentlyPlaying(ctx, accessToken)
	if err != nil {
		if _, ok := err.(*spotify.RateLimitError); !ok {
			log.Printf("fetch currently playing for user %s: %v", userID, err)
		}
		current = nil
	}

	events := make([]db.ListeningEvent, 0, len(recentItems)+1)
	artistIDs := map[string]struct{}{}
	for _, item := range recentItems {
		events = append(events, toListeningEvent(userID, item))
		if item.ArtistID != "" {
			artistIDs[item.ArtistID] = struct{}{}
		}
	}
	if current != nil {
		events = append(events, toListeningEvent(userID, *current))
		if current.ArtistID != "" {
			artistIDs[current.ArtistID] = struct{}{}
		}
	}

	if _, err := db.InsertListeningEvents(ctx, p.Pool, events); err != nil {
		return fmt.Errorf("insert listening events: %w", err)
	}

	if len(artistIDs) > 0 {
		ids := make([]string, 0, len(artistIDs))
		for id := range artistIDs {
			ids = append(ids, id)
		}
		if err := p.refreshGenres(ctx, accessToken, ids); err != nil {
			log.Printf("refresh genres for user %s: %v", userID, err)
		}
	}
	return nil
}

func (p *Poller) refreshGenres(ctx context.Context, accessToken string, ids []string) error {
	staleIDs, err := db.MissingOrStaleArtistIDs(ctx, p.Pool, ids, genreStaleAfter)
	if err != nil {
		return fmt.Errorf("find stale artist ids: %w", err)
	}
	if len(staleIDs) == 0 {
		return nil
	}

	artistGenres, err := p.Spotify.ArtistsByIDs(ctx, accessToken, staleIDs)
	if err != nil {
		return fmt.Errorf("fetch artist genres: %w", err)
	}
	for _, ag := range artistGenres {
		if err := db.UpsertArtistGenres(ctx, p.Pool, ag.ArtistID, ag.Genres); err != nil {
			return fmt.Errorf("upsert artist genres: %w", err)
		}
	}
	return nil
}

func toListeningEvent(userID uuid.UUID, item spotify.ListeningItem) db.ListeningEvent {
	return db.ListeningEvent{
		UserID:     userID,
		TrackID:    item.TrackID,
		TrackName:  item.TrackName,
		ArtistID:   item.ArtistID,
		ArtistName: item.ArtistName,
		PlayedAt:   item.PlayedAt,
		Source:     item.Source,
	}
}
