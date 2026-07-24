package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListeningEvent struct {
	UserID     uuid.UUID
	TrackID    string
	TrackName  string
	ArtistID   string
	ArtistName string
	PlayedAt   time.Time
	Source     string
}

// InsertListeningEvents batch-inserts events, relying on the
// (user_id, track_id, played_at, source) unique constraint plus
// ON CONFLICT DO NOTHING to make overlapping poll windows idempotent.
func InsertListeningEvents(ctx context.Context, pool *pgxpool.Pool, events []ListeningEvent) (inserted int64, err error) {
	if len(events) == 0 {
		return 0, nil
	}

	batch := &pgx.Batch{}
	for _, e := range events {
		batch.Queue(`
			INSERT INTO listening_events (user_id, track_id, track_name, artist_id, artist_name, played_at, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT DO NOTHING
		`, e.UserID, e.TrackID, e.TrackName, e.ArtistID, e.ArtistName, e.PlayedAt, e.Source)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for range events {
		tag, err := br.Exec()
		if err != nil {
			return inserted, fmt.Errorf("insert listening event: %w", err)
		}
		inserted += tag.RowsAffected()
	}
	return inserted, nil
}

// LastPlayedAt returns the most recent recently-played timestamp on record
// for a user, so the poller only asks Spotify for what's new. The second
// return value is false if the user has no recorded plays yet.
func LastPlayedAt(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (time.Time, bool, error) {
	var t *time.Time
	err := pool.QueryRow(ctx, `
		SELECT max(played_at) FROM listening_events
		WHERE user_id = $1 AND source = 'recently_played'
	`, userID).Scan(&t)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("last played at: %w", err)
	}
	if t == nil {
		return time.Time{}, false, nil
	}
	return *t, true, nil
}
