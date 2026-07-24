package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const baseURL = "https://api.spotify.com/v1"

type Client struct {
	httpClient *http.Client
}

func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient}
}

// get performs an authenticated GET and decodes a JSON body into out.
// A 204 response is reported via noContent=true with out left untouched,
// letting callers (currently-playing, notably) distinguish "nothing playing"
// from an error.
func (c *Client) get(ctx context.Context, accessToken, path string, query url.Values, out any) (noContent bool, err error) {
	full := baseURL + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
	if err != nil {
		return false, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return true, nil
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return false, &RateLimitError{RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return false, &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	if out == nil {
		return false, nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}
	return false, nil
}
