package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ Cache = (*InMemoryCache)(nil)

func TestInMemoryCache_GetMissingKey(t *testing.T) {
	cache := NewInMemoryCache()

	value, ok := cache.Get("missing")

	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestInMemoryCache_SetAndGet(t *testing.T) {
	cache := NewInMemoryCache()
	want := map[string]int{"count": 3}

	cache.Set("key", want, time.Minute)
	value, ok := cache.Get("key")

	require.True(t, ok)
	assert.Equal(t, want, value)
}

func TestInMemoryCache_NonPositiveTTLDoesNotExpire(t *testing.T) {
	for _, ttl := range []time.Duration{0, -time.Second} {
		t.Run(fmt.Sprintf("ttl_%s", ttl), func(t *testing.T) {
			cache := NewInMemoryCache()
			cache.Set("key", "value", ttl)

			time.Sleep(10 * time.Millisecond)
			value, ok := cache.Get("key")

			assert.True(t, ok)
			assert.Equal(t, "value", value)
		})
	}
}

func TestInMemoryCache_GetExpiredEntryEvictsIt(t *testing.T) {
	cache := NewInMemoryCache()
	cache.Set("key", "value", time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	value, ok := cache.Get("key")

	assert.False(t, ok)
	assert.Nil(t, value)

	cache.mu.RLock()
	_, exists := cache.entries["key"]
	cache.mu.RUnlock()
	assert.False(t, exists)
}