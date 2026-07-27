package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

var ErrUserNotFound = errors.New("user not found")

// GetUser loads a user by id, used by the /api/me handler to answer "who is
// this session for" without stashing display_name in the session cookie
// itself.
func GetUser(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*User, error) {
	user := &User{ID: id}
	err := pool.QueryRow(ctx, `
		SELECT spotify_id, display_name, created_at FROM users WHERE id = $1
	`, id).Scan(&user.SpotifyID, &user.DisplayName, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
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
