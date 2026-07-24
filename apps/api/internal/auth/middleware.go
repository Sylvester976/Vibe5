package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// RequireSession wraps a handler so it only runs for requests carrying a
// valid session cookie, making the authenticated user id available via
// UserIDFromContext.
func RequireSession(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := VerifySession(r, secret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return id, ok
}
