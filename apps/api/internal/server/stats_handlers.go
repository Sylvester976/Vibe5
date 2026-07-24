package server

import (
	"log"
	"net/http"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/stats"
)

// periodFromQuery reads ?period=, defaulting to medium_term and rejecting
// anything outside the three valid values.
func periodFromQuery(r *http.Request) (string, bool) {
	period := r.URL.Query().Get("period")
	if period == "" {
		return stats.DefaultPeriod, true
	}
	return period, stats.ValidPeriods[period]
}

func (s *Server) handleTopArtists(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	period, ok := periodFromQuery(r)
	if !ok {
		http.Error(w, "invalid period", http.StatusBadRequest)
		return
	}
	resp, err := stats.TopArtists(r.Context(), s.pool, userID, period)
	if err != nil {
		log.Println("top artists:", err)
		http.Error(w, "failed to load top artists", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGenres(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	period, ok := periodFromQuery(r)
	if !ok {
		http.Error(w, "invalid period", http.StatusBadRequest)
		return
	}
	resp, err := stats.Genres(r.Context(), s.pool, userID, period)
	if err != nil {
		log.Println("genres:", err)
		http.Error(w, "failed to load genres", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleHeatmap(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	resp, err := stats.Heatmap(r.Context(), s.pool, userID)
	if err != nil {
		log.Println("heatmap:", err)
		http.Error(w, "failed to load heatmap", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
