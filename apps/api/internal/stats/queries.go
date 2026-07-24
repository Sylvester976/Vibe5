// Package stats reads the worker's rollup output (top_snapshots) plus raw
// listening_events (for the heatmap, which has no snapshot table) and shapes
// it into the dashboard/story DTOs the API contract promises.
package stats

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"spotify-insights/internal/db"
)

// TopArtists reads the latest snapshot for a period. Returns an
// empty-but-well-formed response if the worker hasn't rolled up yet.
func TopArtists(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, period string) (*TopArtistsResponse, error) {
	snap, err := db.LatestTopSnapshot(ctx, pool, userID, period)
	if err != nil {
		return nil, fmt.Errorf("load top artists: %w", err)
	}
	resp := &TopArtistsResponse{Period: period}
	if snap != nil {
		resp.SnapshotDate = snap.SnapshotDate.Format("2006-01-02")
		resp.Artists = snap.TopArtists
	}
	// A nil slice here (no snapshot yet, or a snapshot with zero plays for
	// the period) would marshal to JSON `null` instead of `[]`, which trips
	// up a JS/React client calling .map() on it.
	if resp.Artists == nil {
		resp.Artists = []db.ArtistCount{}
	}
	return resp, nil
}

// Genres derives a genre distribution by joining the period's top-artists
// snapshot against the cached artist_genres table, weighting each genre by
// its artist's play count.
func Genres(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, period string) (*GenresResponse, error) {
	snap, err := db.LatestTopSnapshot(ctx, pool, userID, period)
	if err != nil {
		return nil, fmt.Errorf("load genres: %w", err)
	}
	resp := &GenresResponse{Period: period, Genres: []GenreCount{}}
	if snap == nil {
		return resp, nil
	}

	ids := make([]string, 0, len(snap.TopArtists))
	for _, a := range snap.TopArtists {
		ids = append(ids, a.ArtistID)
	}
	genresByArtist, err := db.GenresForArtists(ctx, pool, ids)
	if err != nil {
		return nil, fmt.Errorf("load artist genres: %w", err)
	}

	counts := map[string]int{}
	for _, a := range snap.TopArtists {
		for _, g := range genresByArtist[a.ArtistID] {
			counts[g] += a.PlayCount
		}
	}

	resp.Genres = make([]GenreCount, 0, len(counts))
	for g, c := range counts {
		resp.Genres = append(resp.Genres, GenreCount{Genre: g, Count: c})
	}
	sort.Slice(resp.Genres, func(i, j int) bool { return resp.Genres[i].Count > resp.Genres[j].Count })
	return resp, nil
}

// Heatmap reads raw listening_events directly — there's no snapshot table
// for it, so it always reflects full recorded history.
func Heatmap(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*HeatmapResponse, error) {
	rows, err := pool.Query(ctx, `
		SELECT extract(dow from played_at)::int AS day_of_week,
		       extract(hour from played_at)::int AS hour,
		       count(*)::int
		FROM listening_events
		WHERE user_id = $1
		GROUP BY day_of_week, hour
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query heatmap: %w", err)
	}
	defer rows.Close()

	resp := &HeatmapResponse{Cells: []HeatmapCell{}}
	for rows.Next() {
		var cell HeatmapCell
		if err := rows.Scan(&cell.DayOfWeek, &cell.Hour, &cell.Count); err != nil {
			return nil, fmt.Errorf("scan heatmap cell: %w", err)
		}
		resp.Cells = append(resp.Cells, cell)
	}
	return resp, rows.Err()
}

// Story composes the other three queries plus a total-plays count into an
// ordered slide list for the story view.
func Story(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*StoryResponse, error) {
	topArtists, err := TopArtists(ctx, pool, userID, "long_term")
	if err != nil {
		return nil, err
	}
	genres, err := Genres(ctx, pool, userID, "long_term")
	if err != nil {
		return nil, err
	}
	heatmap, err := Heatmap(ctx, pool, userID)
	if err != nil {
		return nil, err
	}

	var totalPlays int
	for _, c := range heatmap.Cells {
		totalPlays += c.Count
	}

	slides := []StorySlide{
		{Type: "welcome", Title: "Your Vibe", Caption: "Here's your listening story so far."},
	}
	if len(topArtists.Artists) > 0 {
		top := topArtists.Artists[0]
		slides = append(slides, StorySlide{
			Type:    "top_artist",
			Title:   "Top Artist",
			Value:   top.ArtistName,
			Caption: fmt.Sprintf("%d plays and counting", top.PlayCount),
		})
	}
	if len(genres.Genres) > 0 {
		slides = append(slides, StorySlide{
			Type:  "top_genre",
			Title: "Top Genre",
			Value: genres.Genres[0].Genre,
		})
	}
	slides = append(slides, StorySlide{
		Type:  "total_plays",
		Title: "Total Plays Recorded",
		Value: fmt.Sprintf("%d", totalPlays),
	})
	slides = append(slides, StorySlide{
		Type:    "goodbye",
		Title:   "That's Your Vibe",
		Caption: "Check back as we keep listening.",
	})

	return &StoryResponse{Slides: slides}, nil
}
