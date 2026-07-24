package stats

import "spotify-insights/internal/db"

var ValidPeriods = map[string]bool{"short_term": true, "medium_term": true, "long_term": true}

const DefaultPeriod = "medium_term"

type TopArtistsResponse struct {
	Period       string           `json:"period"`
	SnapshotDate string           `json:"snapshot_date,omitempty"`
	Artists      []db.ArtistCount `json:"artists"`
}

type GenreCount struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

type GenresResponse struct {
	Period string       `json:"period"`
	Genres []GenreCount `json:"genres"`
}

// HeatmapCell.DayOfWeek follows Postgres's extract(dow ...): 0 = Sunday .. 6 = Saturday.
type HeatmapCell struct {
	DayOfWeek int `json:"day_of_week"`
	Hour      int `json:"hour"`
	Count     int `json:"count"`
}

type HeatmapResponse struct {
	Cells []HeatmapCell `json:"cells"`
}

type StorySlide struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Value   string `json:"value,omitempty"`
	Caption string `json:"caption,omitempty"`
}

type StoryResponse struct {
	Slides []StorySlide `json:"slides"`
}
