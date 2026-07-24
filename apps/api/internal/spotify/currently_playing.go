package spotify

import (
	"context"
	"time"
)

// CurrentlyPlaying returns the track playing right now, or nil if nothing is
// playing (Spotify reports that as a 204 with no body).
func (c *Client) CurrentlyPlaying(ctx context.Context, accessToken string) (*ListeningItem, error) {
	var raw rawCurrentlyPlayingResponse
	noContent, err := c.get(ctx, accessToken, "/me/player/currently-playing", nil, &raw)
	if err != nil {
		return nil, err
	}
	if noContent || !raw.IsPlaying || raw.Item.ID == "" {
		return nil, nil
	}

	artistID, artistName := primaryArtist(raw.Item)
	return &ListeningItem{
		TrackID:    raw.Item.ID,
		TrackName:  raw.Item.Name,
		ArtistID:   artistID,
		ArtistName: artistName,
		PlayedAt:   time.Now().UTC(),
		Source:     "currently_playing",
	}, nil
}
