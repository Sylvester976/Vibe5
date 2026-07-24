package spotify

import (
	"context"
	"net/url"
	"strings"
)

const artistsChunkSize = 50

// ArtistsByIDs fetches genres for up to any number of artist ids, chunking
// requests into batches of 50 (Spotify's per-request limit for GET /artists).
func (c *Client) ArtistsByIDs(ctx context.Context, accessToken string, ids []string) ([]ArtistGenres, error) {
	var results []ArtistGenres
	for start := 0; start < len(ids); start += artistsChunkSize {
		end := min(start+artistsChunkSize, len(ids))
		chunk := ids[start:end]

		var raw rawArtistsResponse
		query := url.Values{"ids": {strings.Join(chunk, ",")}}
		if _, err := c.get(ctx, accessToken, "/artists", query, &raw); err != nil {
			return nil, err
		}
		for _, a := range raw.Artists {
			results = append(results, ArtistGenres{ArtistID: a.ID, Genres: a.Genres})
		}
	}
	return results, nil
}
