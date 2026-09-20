package auth

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RoleSource resolves the current effective roles of a user from the
// authoritative store. It lets session resolution reflect role changes without
// forcing a re-login.
type RoleSource interface {
	RolesForUser(ctx context.Context, userID uuid.UUID) ([]string, error)
}

const (
	defaultRoleCacheTTL = 30 * time.Second
	maxRoleCacheEntries = 10_000
)

// NewCachingRoleSource wraps a RoleSource with a short-lived cache to bound the
// number of lookups per request. A nil inner source returns nil.
func NewCachingRoleSource(inner RoleSource, ttl time.Duration) RoleSource {
	if inner == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = defaultRoleCacheTTL
	}
	return &cachingRoleSource{
		inner:   inner,
		ttl:     ttl,
		now:     time.Now,
		entries: make(map[uuid.UUID]roleCacheEntry),
	}
}

type cachingRoleSource struct {
	inner RoleSource
	ttl   time.Duration
	now   func() time.Time

	mu      sync.Mutex
	entries map[uuid.UUID]roleCacheEntry
}

type roleCacheEntry struct {
	roles   []string
	expires time.Time
}

// RolesForUser implements RoleSource.
func (c *cachingRoleSource) RolesForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	c.mu.Lock()
	if entry, ok := c.entries[userID]; ok && c.now().Before(entry.expires) {
		roles := append([]string(nil), entry.roles...)
		c.mu.Unlock()
		return roles, nil
	}
	c.mu.Unlock()

	roles, err := c.inner.RolesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	if len(c.entries) >= maxRoleCacheEntries {
		c.evictLocked()
	}
	c.entries[userID] = roleCacheEntry{
		roles:   append([]string(nil), roles...),
		expires: c.now().Add(c.ttl),
	}
	c.mu.Unlock()
	return append([]string(nil), roles...), nil
}

func (c *cachingRoleSource) evictLocked() {
	now := c.now()
	for key, entry := range c.entries {
		if !now.Before(entry.expires) {
			delete(c.entries, key)
		}
	}
	if len(c.entries) >= maxRoleCacheEntries {
		for key := range c.entries {
			delete(c.entries, key)
			break
		}
	}
}
