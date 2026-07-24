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

var ErrWidgetTokenNotFound = errors.New("widget token not found")

type WidgetTokenRow struct {
	Token     string
	UserID    uuid.UUID
	CreatedAt time.Time
	Revoked   bool
}

func InsertWidgetToken(ctx context.Context, pool *pgxpool.Pool, token string, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `INSERT INTO widget_tokens (token, user_id) VALUES ($1, $2)`, token, userID)
	if err != nil {
		return fmt.Errorf("insert widget token: %w", err)
	}
	return nil
}

// RevokeActiveWidgetTokens revokes every non-revoked token for a user;
// "rotate" (POST /api/widget/token) is this followed by InsertWidgetToken.
func RevokeActiveWidgetTokens(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `UPDATE widget_tokens SET revoked = true WHERE user_id = $1 AND NOT revoked`, userID)
	if err != nil {
		return fmt.Errorf("revoke widget tokens: %w", err)
	}
	return nil
}

func GetWidgetToken(ctx context.Context, pool *pgxpool.Pool, token string) (*WidgetTokenRow, error) {
	row := &WidgetTokenRow{Token: token}
	err := pool.QueryRow(ctx, `
		SELECT user_id, created_at, revoked FROM widget_tokens WHERE token = $1
	`, token).Scan(&row.UserID, &row.CreatedAt, &row.Revoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWidgetTokenNotFound
		}
		return nil, fmt.Errorf("get widget token: %w", err)
	}
	return row, nil
}
