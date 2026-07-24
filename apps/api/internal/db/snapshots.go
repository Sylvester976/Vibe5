package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArtistCount struct {
	ArtistID   string `json:"artist_id"`
	ArtistName string `json:"artist_name"`
	PlayCount  int    `json:"play_count"`
}

type TrackCount struct {
	TrackID    string `json:"track_id"`
	TrackName  string `json:"track_name"`
	ArtistName string `json:"artist_name"`
	PlayCount  int    `json:"play_count"`
}

const topSnapshotLimit = 20

// TopArtistCounts aggregates raw listening_events for one user since a
// cutoff time, most-played first. Used by the nightly rollup — snapshots
// are a proxy over our own collected history, not a mirror of Spotify's own
// top-items ranking.
func TopArtistCounts(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, since time.Time) ([]ArtistCount, error) {
	rows, err := pool.Query(ctx, `
		SELECT artist_id, max(artist_name) AS artist_name, count(*) AS play_count
		FROM listening_events
		WHERE user_id = $1 AND played_at >= $2
		GROUP BY artist_id
		ORDER BY play_count DESC
		LIMIT $3
	`, userID, since, topSnapshotLimit)
	if err != nil {
		return nil, fmt.Errorf("query top artists: %w", err)
	}
	defer rows.Close()

	out := []ArtistCount{}
	for rows.Next() {
		var a ArtistCount
		if err := rows.Scan(&a.ArtistID, &a.ArtistName, &a.PlayCount); err != nil {
			return nil, fmt.Errorf("scan top artist: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// TopTrackCounts is TopArtistCounts' track-level equivalent.
func TopTrackCounts(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, since time.Time) ([]TrackCount, error) {
	rows, err := pool.Query(ctx, `
		SELECT track_id, max(track_name) AS track_name, max(artist_name) AS artist_name, count(*) AS play_count
		FROM listening_events
		WHERE user_id = $1 AND played_at >= $2
		GROUP BY track_id
		ORDER BY play_count DESC
		LIMIT $3
	`, userID, since, topSnapshotLimit)
	if err != nil {
		return nil, fmt.Errorf("query top tracks: %w", err)
	}
	defer rows.Close()

	out := []TrackCount{}
	for rows.Next() {
		var t TrackCount
		if err := rows.Scan(&t.TrackID, &t.TrackName, &t.ArtistName, &t.PlayCount); err != nil {
			return nil, fmt.Errorf("scan top track: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func UpsertTopSnapshot(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, period string, snapshotDate time.Time, topArtists []ArtistCount, topTracks []TrackCount) error {
	artistsJSON, err := json.Marshal(topArtists)
	if err != nil {
		return fmt.Errorf("marshal top artists: %w", err)
	}
	tracksJSON, err := json.Marshal(topTracks)
	if err != nil {
		return fmt.Errorf("marshal top tracks: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO top_snapshots (user_id, period, snapshot_date, top_artists, top_tracks)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, period, snapshot_date) DO UPDATE SET
			top_artists = EXCLUDED.top_artists,
			top_tracks = EXCLUDED.top_tracks
	`, userID, period, snapshotDate, artistsJSON, tracksJSON)
	if err != nil {
		return fmt.Errorf("upsert top snapshot: %w", err)
	}
	return nil
}

type TopSnapshot struct {
	SnapshotDate time.Time
	TopArtists   []ArtistCount
	TopTracks    []TrackCount
}

// LatestTopSnapshot returns the most recent snapshot for user+period, or nil
// (not an error) if the worker hasn't produced one yet.
func LatestTopSnapshot(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, period string) (*TopSnapshot, error) {
	var (
		snap                  TopSnapshot
		artistsRaw, tracksRaw []byte
	)
	err := pool.QueryRow(ctx, `
		SELECT snapshot_date, top_artists, top_tracks FROM top_snapshots
		WHERE user_id = $1 AND period = $2
		ORDER BY snapshot_date DESC
		LIMIT 1
	`, userID, period).Scan(&snap.SnapshotDate, &artistsRaw, &tracksRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query latest top snapshot: %w", err)
	}

	if err := json.Unmarshal(artistsRaw, &snap.TopArtists); err != nil {
		return nil, fmt.Errorf("unmarshal top artists: %w", err)
	}
	if err := json.Unmarshal(tracksRaw, &snap.TopTracks); err != nil {
		return nil, fmt.Errorf("unmarshal top tracks: %w", err)
	}
	return &snap, nil
}
