package server

import (
	"errors"
	"log"
	"net/http"
	"time"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/db"
	"spotify-insights/internal/widget"
)

const widgetCacheTTL = 15 * time.Second

// handleIssueWidgetToken rotates the caller's widget token: any previously
// issued token is revoked and a new one issued, so sharing a fresh embed URL
// invalidates any old one still floating around.
func (s *Server) handleIssueWidgetToken(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	if err := db.RevokeActiveWidgetTokens(r.Context(), s.pool, userID); err != nil {
		log.Println("revoke widget tokens:", err)
		http.Error(w, "failed to rotate widget token", http.StatusInternalServerError)
		return
	}

	token := widget.NewToken(s.cfg.WidgetHMACSecret, userID)
	if err := db.InsertWidgetToken(r.Context(), s.pool, token, userID); err != nil {
		log.Println("insert widget token:", err)
		http.Error(w, "failed to issue widget token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"widget_url": "/widget/" + token})
}

// handleWidget is public and unauthenticated by design (it's meant to be
// embedded in someone else's README/site). The signature is verified before
// anything else touches the database or Spotify, so a malformed/forged token
// costs nothing beyond the HMAC check.
func (s *Server) handleWidget(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	userID, ok := widget.Verify(s.cfg.WidgetHMACSecret, token)
	if !ok {
		http.Error(w, "invalid widget token", http.StatusBadRequest)
		return
	}

	if svg, ok := s.widgetCache.Get(token); ok {
		writeSVG(w, svg)
		return
	}

	row, err := db.GetWidgetToken(r.Context(), s.pool, token)
	if err != nil {
		if errors.Is(err, db.ErrWidgetTokenNotFound) {
			http.Error(w, "widget not found", http.StatusNotFound)
			return
		}
		log.Println("get widget token:", err)
		http.Error(w, "failed to load widget", http.StatusInternalServerError)
		return
	}
	if row.Revoked {
		http.Error(w, "widget token revoked", http.StatusGone)
		return
	}

	accessToken, err := s.tokens.GetValidAccessToken(r.Context(), userID)
	if err != nil {
		log.Println("get access token for widget:", err)
		http.Error(w, "failed to load now playing", http.StatusBadGateway)
		return
	}
	current, err := s.spotify.CurrentlyPlaying(r.Context(), accessToken)
	if err != nil {
		log.Println("fetch currently playing for widget:", err)
		http.Error(w, "failed to load now playing", http.StatusBadGateway)
		return
	}

	svg := widget.RenderNowPlayingSVG(current)
	s.widgetCache.Set(token, svg, widgetCacheTTL)
	writeSVG(w, svg)
}

func writeSVG(w http.ResponseWriter, svg []byte) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(svg)
}
