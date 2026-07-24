package poller

import (
	"context"
	"fmt"
	"log"
	"time"

	"spotify-insights/internal/db"
)

type rollupPeriod struct {
	Name     string
	Lookback time.Duration // 0 means all-time
}

// Periods are windows over our own collected listening_events, not a mirror
// of Spotify's own (opaque) top-items ranking — short/medium/long_term here
// mean "last 28 days / 180 days / all recorded history".
var rollupPeriods = []rollupPeriod{
	{Name: "short_term", Lookback: 28 * 24 * time.Hour},
	{Name: "medium_term", Lookback: 180 * 24 * time.Hour},
	{Name: "long_term", Lookback: 0},
}

// allTimeLookback stands in for "no lower bound" in a query that always
// takes a since time.
const allTimeLookback = 20 * 365 * 24 * time.Hour

// RunNightlyRollup aggregates every user's listening_events into a
// top_snapshots row per period, so dashboard/story reads never scan raw
// event rows.
func (p *Poller) RunNightlyRollup(ctx context.Context) error {
	userIDs, err := db.ListUserIDs(ctx, p.Pool)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	now := time.Now().UTC()
	snapshotDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	for _, userID := range userIDs {
		for _, per := range rollupPeriods {
			lookback := per.Lookback
			if lookback == 0 {
				lookback = allTimeLookback
			}
			since := snapshotDate.Add(-lookback)

			topArtists, err := db.TopArtistCounts(ctx, p.Pool, userID, since)
			if err != nil {
				log.Printf("rollup top artists user=%s period=%s: %v", userID, per.Name, err)
				continue
			}
			topTracks, err := db.TopTrackCounts(ctx, p.Pool, userID, since)
			if err != nil {
				log.Printf("rollup top tracks user=%s period=%s: %v", userID, per.Name, err)
				continue
			}
			if err := db.UpsertTopSnapshot(ctx, p.Pool, userID, per.Name, snapshotDate, topArtists, topTracks); err != nil {
				log.Printf("upsert snapshot user=%s period=%s: %v", userID, per.Name, err)
			}
		}
	}
	return nil
}
