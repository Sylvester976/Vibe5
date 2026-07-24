package spotify

import "time"

type Profile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ListeningItem is our normalized shape for both a recently-played entry and
// a currently-playing snapshot — the two sources the poller writes to
// listening_events, distinguished by Source.
type ListeningItem struct {
	TrackID    string
	TrackName  string
	ArtistID   string
	ArtistName string
	PlayedAt   time.Time
	Source     string // "recently_played" | "currently_playing"
}

type ArtistGenres struct {
	ArtistID string
	Genres   []string
}

type rawTrack struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Artists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
}

type rawRecentlyPlayedResponse struct {
	Items []struct {
		Track    rawTrack  `json:"track"`
		PlayedAt time.Time `json:"played_at"`
	} `json:"items"`
}

type rawCurrentlyPlayingResponse struct {
	IsPlaying bool     `json:"is_playing"`
	Item      rawTrack `json:"item"`
}

type rawArtistsResponse struct {
	Artists []struct {
		ID     string   `json:"id"`
		Genres []string `json:"genres"`
	} `json:"artists"`
}
