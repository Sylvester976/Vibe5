package spotify

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/db"
)

// refreshSkew is how much lead time before expiry we proactively refresh,
// so a request never races an access token expiring mid-flight.
const refreshSkew = 60 * time.Second

// TokenManager is the single place that reads/writes encrypted Spotify
// tokens and knows how to refresh them. Both the poller and the widget
// endpoint depend on GetValidAccessToken to get a live, unexpired token.
type TokenManager struct {
	Pool          *pgxpool.Pool
	OAuthConfig   *oauth2.Config
	EncryptionKey []byte
}

func NewTokenManager(pool *pgxpool.Pool, oauthCfg *oauth2.Config, encryptionKey []byte) *TokenManager {
	return &TokenManager{Pool: pool, OAuthConfig: oauthCfg, EncryptionKey: encryptionKey}
}

// StoreTokens encrypts and persists a freshly obtained OAuth token pair,
// used right after the initial /auth/callback exchange.
func (tm *TokenManager) StoreTokens(ctx context.Context, userID uuid.UUID, accessToken, refreshToken string, expiry time.Time) error {
	accessCipher, err := auth.Encrypt(tm.EncryptionKey, accessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	refreshCipher, err := auth.Encrypt(tm.EncryptionKey, refreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}
	return db.UpsertSpotifyTokens(ctx, tm.Pool, userID, accessCipher, refreshCipher, expiry)
}

// GetValidAccessToken returns a usable access token for userID, transparently
// refreshing it against Spotify if it's expired or about to expire.
func (tm *TokenManager) GetValidAccessToken(ctx context.Context, userID uuid.UUID) (string, error) {
	row, err := db.GetSpotifyTokens(ctx, tm.Pool, userID)
	if err != nil {
		return "", fmt.Errorf("load tokens: %w", err)
	}
	if time.Now().Add(refreshSkew).Before(row.ExpiresAt) {
		return auth.Decrypt(tm.EncryptionKey, row.AccessTokenCipher)
	}
	return tm.refresh(ctx, userID, row)
}

func (tm *TokenManager) refresh(ctx context.Context, userID uuid.UUID, row *db.SpotifyTokenRow) (string, error) {
	refreshToken, err := auth.Decrypt(tm.EncryptionKey, row.RefreshTokenCipher)
	if err != nil {
		return "", fmt.Errorf("decrypt refresh token: %w", err)
	}

	tok, err := tm.OAuthConfig.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		return "", fmt.Errorf("refresh token: %w", err)
	}

	// Spotify does not always issue a new refresh token on refresh — keep
	// the existing one when it doesn't.
	newRefreshToken := tok.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = refreshToken
	}

	if err := tm.StoreTokens(ctx, userID, tok.AccessToken, newRefreshToken, tok.Expiry); err != nil {
		return "", fmt.Errorf("persist refreshed tokens: %w", err)
	}
	return tok.AccessToken, nil
}
