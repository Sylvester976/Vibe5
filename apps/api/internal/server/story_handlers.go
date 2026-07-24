package server

import (
	"log"
	"net/http"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/stats"
)

func (s *Server) handleStory(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	resp, err := stats.Story(r.Context(), s.pool, userID)
	if err != nil {
		log.Println("story:", err)
		http.Error(w, "failed to load story", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
