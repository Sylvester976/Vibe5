package widget

import (
	"sync"
	"time"
)

type cacheEntry struct {
	svg       []byte
	expiresAt time.Time
}

// Cache is a short-TTL, in-memory cache for rendered widget SVGs, keyed by
// widget token. Chosen over Redis per the project's scale (Spotify Dev Mode
// caps at 5 authorized users) — its purpose is only to shield the
// unauthenticated /widget/{token} endpoint from triggering a live Spotify
// call on every embed page-load.
type Cache struct {
	m sync.Map
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Get(token string) ([]byte, bool) {
	v, ok := c.m.Load(token)
	if !ok {
		return nil, false
	}
	entry := v.(cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.m.Delete(token)
		return nil, false
	}
	return entry.svg, true
}

func (c *Cache) Set(token string, svg []byte, ttl time.Duration) {
	c.m.Store(token, cacheEntry{svg: svg, expiresAt: time.Now().Add(ttl)})
}
