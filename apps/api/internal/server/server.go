// Package server wires HTTP routes to handlers — the seam between the
// domain packages (auth, spotify, stats, widget) and net/http that the
// architecture doc's package list doesn't otherwise name.
package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"spotify-insights/internal/auth"
	"spotify-insights/internal/config"
	"spotify-insights/internal/spotify"
	"spotify-insights/internal/widget"
)

type Server struct {
	cfg         *config.Config
	pool        *pgxpool.Pool
	oauthConfig *oauth2.Config
	spotify     *spotify.Client
	tokens      *spotify.TokenManager
	widgetCache *widget.Cache
	mux         *http.ServeMux
}

func New(cfg *config.Config, pool *pgxpool.Pool, spClient *spotify.Client, oauthCfg *oauth2.Config, tokens *spotify.TokenManager) *Server {
	s := &Server{
		cfg:         cfg,
		pool:        pool,
		oauthConfig: oauthCfg,
		spotify:     spClient,
		tokens:      tokens,
		widgetCache: widget.NewCache(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	mux := http.NewServeMux()
	requireSession := auth.RequireSession(s.cfg.SessionSecret)

	mux.HandleFunc("GET /auth/login", s.handleLogin)
	mux.HandleFunc("GET /auth/callback", s.handleCallback)
	mux.Handle("POST /auth/refresh", requireSession(http.HandlerFunc(s.handleRefresh)))
	mux.Handle("GET /api/me", requireSession(http.HandlerFunc(s.handleMe)))

	mux.Handle("GET /api/stats/top-artists", requireSession(http.HandlerFunc(s.handleTopArtists)))
	mux.Handle("GET /api/stats/genres", requireSession(http.HandlerFunc(s.handleGenres)))
	mux.Handle("GET /api/stats/heatmap", requireSession(http.HandlerFunc(s.handleHeatmap)))
	mux.Handle("GET /api/story", requireSession(http.HandlerFunc(s.handleStory)))

	mux.Handle("POST /api/widget/token", requireSession(http.HandlerFunc(s.handleIssueWidgetToken)))
	mux.HandleFunc("GET /widget/{token}", s.handleWidget)

	s.mux = mux
}
