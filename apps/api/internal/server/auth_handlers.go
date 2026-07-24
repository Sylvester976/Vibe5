package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/db"
)

const sessionTTL = 30 * 24 * time.Hour

// handleLogin redirects the browser to Spotify's OAuth consent screen with a
// PKCE challenge and a CSRF state, both stashed in a short-lived cookie.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	state, verifier, err := auth.BeginOAuthState(w, s.cfg.SessionSecret)
	if err != nil {
		log.Println("begin oauth state:", err)
		http.Error(w, "failed to start login", http.StatusInternalServerError)
		return
	}
	authURL := s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))
	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleCallback exchanges the authorization code, links/creates the user,
// stores encrypted tokens, and issues a session cookie. There is no frontend
// yet, so it responds with a small JSON confirmation instead of redirecting
// into a React route.
func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Error(w, "spotify authorization denied: "+errParam, http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	verifier, err := auth.VerifyOAuthState(w, r, s.cfg.SessionSecret, r.URL.Query().Get("state"))
	if err != nil {
		log.Println("verify oauth state:", err)
		http.Error(w, "invalid or expired login attempt", http.StatusBadRequest)
		return
	}

	tok, err := s.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		log.Println("exchange code:", err)
		http.Error(w, "failed to exchange authorization code", http.StatusBadGateway)
		return
	}

	profile, err := s.spotify.Me(ctx, tok.AccessToken)
	if err != nil {
		log.Println("fetch profile:", err)
		http.Error(w, "failed to fetch spotify profile", http.StatusBadGateway)
		return
	}

	userID, err := db.UpsertUser(ctx, s.pool, profile.ID, profile.DisplayName)
	if err != nil {
		log.Println("upsert user:", err)
		http.Error(w, "failed to save user", http.StatusInternalServerError)
		return
	}

	if err := s.tokens.StoreTokens(ctx, userID, tok.AccessToken, tok.RefreshToken, tok.Expiry); err != nil {
		log.Println("store tokens:", err)
		http.Error(w, "failed to save spotify tokens", http.StatusInternalServerError)
		return
	}

	if err := auth.IssueSession(w, s.cfg.SessionSecret, userID, sessionTTL); err != nil {
		log.Println("issue session:", err)
		http.Error(w, "failed to start session", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":       "connected",
		"display_name": profile.DisplayName,
	})
}

// handleRefresh ensures the caller's stored access token is valid, refreshing
// it against Spotify if it's expired or close to expiring.
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := s.tokens.GetValidAccessToken(r.Context(), userID); err != nil {
		log.Println("refresh token:", err)
		http.Error(w, "failed to refresh token", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "refreshed"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("write json response:", err)
	}
}
