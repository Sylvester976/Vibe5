package spotify

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// RecentlyPlayed fetches up to 50 tracks played after the given time.
func (c *Client) RecentlyPlayed(ctx context.Context, accessToken string, after time.Time) ([]ListeningItem, error) {
	query := url.Values{
		"limit": {"50"},
		"after": {strconv.FormatInt(after.UnixMilli(), 10)},
	}

	var raw rawRecentlyPlayedResponse
	if _, err := c.get(ctx, accessToken, "/me/player/recently-played", query, &raw); err != nil {
		return nil, err
	}

	items := make([]ListeningItem, 0, len(raw.Items))
	for _, it := range raw.Items {
		if it.Track.ID == "" {
			continue
		}
		artistID, artistName := primaryArtist(it.Track)
		items = append(items, ListeningItem{
			TrackID:    it.Track.ID,
			TrackName:  it.Track.Name,
			ArtistID:   artistID,
			ArtistName: artistName,
			PlayedAt:   it.PlayedAt,
			Source:     "recently_played",
		})
	}
	return items, nil
}
