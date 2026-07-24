package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func UpsertArtistGenres(ctx context.Context, pool *pgxpool.Pool, artistID string, genres []string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO artist_genres (artist_id, genres, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (artist_id) DO UPDATE SET genres = EXCLUDED.genres, updated_at = now()
	`, artistID, genres)
	if err != nil {
		return fmt.Errorf("upsert artist genres: %w", err)
	}
	return nil
}

// MissingOrStaleArtistIDs filters ids down to the ones we either have never
// cached genres for, or cached more than staleAfter ago.
func MissingOrStaleArtistIDs(ctx context.Context, pool *pgxpool.Pool, ids []string, staleAfter time.Duration) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := pool.Query(ctx, `
		SELECT a.id FROM unnest($1::text[]) AS a(id)
		LEFT JOIN artist_genres g ON g.artist_id = a.id
		WHERE g.artist_id IS NULL OR g.updated_at < $2
	`, ids, time.Now().Add(-staleAfter))
	if err != nil {
		return nil, fmt.Errorf("query missing/stale artist ids: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan artist id: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// GenresForArtists returns a map of artist_id -> cached genres for whichever
// of the given ids we have entries for.
func GenresForArtists(ctx context.Context, pool *pgxpool.Pool, ids []string) (map[string][]string, error) {
	if len(ids) == 0 {
		return map[string][]string{}, nil
	}
	rows, err := pool.Query(ctx, `
		SELECT artist_id, genres FROM artist_genres WHERE artist_id = ANY($1::text[])
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("query artist genres: %w", err)
	}
	defer rows.Close()

	out := map[string][]string{}
	for rows.Next() {
		var id string
		var genres []string
		if err := rows.Scan(&id, &genres); err != nil {
			return nil, fmt.Errorf("scan artist genres: %w", err)
		}
		out[id] = genres
	}
	return out, rows.Err()
}
