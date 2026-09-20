// Package ttlcache is a tiny in-process cache with a fixed TTL, used for
// public responses that do not need per-user accuracy.
package ttlcache

import (
	"sync"
	"time"
)

type entry[T any] struct {
	value   T
	expires time.Time
}

// Cache is a concurrency-safe TTL cache.
type Cache[T any] struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[string]entry[T]
}

// New creates a cache whose entries live for ttl.
func New[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{ttl: ttl, items: make(map[string]entry[T])}
}

// Get returns a cached value if it is present and unexpired.
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.items[key]
	if !ok || time.Now().After(item.expires) {
		var zero T
		delete(c.items, key)
		return zero, false
	}
	return item.value, true
}

// Set stores a value for the cache's TTL.
func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry[T]{value: value, expires: time.Now().Add(c.ttl)}
}
