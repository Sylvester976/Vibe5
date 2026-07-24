package spotify

import "context"

func (c *Client) Me(ctx context.Context, accessToken string) (*Profile, error) {
	var profile Profile
	if _, err := c.get(ctx, accessToken, "/me", nil, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// primaryArtist safely reads a track's first artist. Some Spotify items
// (e.g. local files, some podcast episodes) can carry an empty artists list.
func primaryArtist(t rawTrack) (id, name string) {
	if len(t.Artists) == 0 {
		return "", ""
	}
	return t.Artists[0].ID, t.Artists[0].Name
}
