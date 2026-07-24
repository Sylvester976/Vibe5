package spotify

import (
	"fmt"
	"strconv"
	"time"
)

// RateLimitError is returned when Spotify responds 429. Callers (the poller)
// should back off the affected user rather than treat this as a hard failure.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("spotify rate limited, retry after %s", e.RetryAfter)
}

// APIError wraps any other non-2xx Spotify response.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("spotify api error: status %d: %s", e.StatusCode, e.Body)
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 5 * time.Second
	}
	if secs, err := strconv.Atoi(header); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 5 * time.Second
}
