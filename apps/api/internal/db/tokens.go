package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SpotifyTokenRow holds the encrypted-at-rest token ciphertext exactly as
// stored — callers decrypt with internal/auth.Decrypt.
type SpotifyTokenRow struct {
	UserID             uuid.UUID
	AccessTokenCipher  string
	RefreshTokenCipher string
	ExpiresAt          time.Time
}

func UpsertSpotifyTokens(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, accessTokenCipher, refreshTokenCipher string, expiresAt time.Time) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO spotify_tokens (user_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at
	`, userID, accessTokenCipher, refreshTokenCipher, expiresAt)
	if err != nil {
		return fmt.Errorf("upsert spotify tokens: %w", err)
	}
	return nil
}

func GetSpotifyTokens(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*SpotifyTokenRow, error) {
	row := &SpotifyTokenRow{UserID: userID}
	err := pool.QueryRow(ctx, `
		SELECT access_token, refresh_token, expires_at
		FROM spotify_tokens WHERE user_id = $1
	`, userID).Scan(&row.AccessTokenCipher, &row.RefreshTokenCipher, &row.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("get spotify tokens: %w", err)
	}
	return row, nil
}
