package cache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
}

type CacheEntry struct {
	value 		any
	expiresAt	time.Time
}

type InMemoryCache struct {
	mu		sync.RWMutex
	entries	map[string]CacheEntry
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[string]CacheEntry),
	}
}

func (c *InMemoryCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time

	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	
	c.entries[key] = CacheEntry{
		value: value,
		expiresAt: expiresAt,
	}
}

func (c *InMemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()

	entry, ok := c.entries[key]
	if !ok {
		c.mu.RUnlock()
		return nil, false
	}

	if entry.expiresAt.IsZero() || time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.value, true
	}

	c.mu.RUnlock()
	return c.evictExpiredCache(key)
}

func (c *InMemoryCache) evictExpiredCache(key string)  (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	currentEntry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if !currentEntry.expiresAt.IsZero() && !time.Now().Before(currentEntry.expiresAt) {
        delete(c.entries, key)
        return nil, false
    }
	return currentEntry.value, true
}