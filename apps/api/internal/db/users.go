package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID          uuid.UUID
	SpotifyID   string
	DisplayName string
	CreatedAt   time.Time
}

// UpsertUser creates a user row for a Spotify account on first login, or
// refreshes the display name on subsequent logins, returning the user's id
// either way.
func UpsertUser(ctx context.Context, pool *pgxpool.Pool, spotifyID, displayName string) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO users (spotify_id, display_name)
		VALUES ($1, $2)
		ON CONFLICT (spotify_id) DO UPDATE SET display_name = EXCLUDED.display_name
		RETURNING id
	`, spotifyID, displayName).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert user: %w", err)
	}
	return id, nil
}

// ListUserIDs returns every user id, used by the worker to iterate all
// connected accounts on each poll/rollup tick.
func ListUserIDs(ctx context.Context, pool *pgxpool.Pool) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `SELECT id FROM users`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
