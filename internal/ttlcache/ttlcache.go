// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package ttlcache

import (
	"context"
	"sync"
	"time"
)

// Cache is a concurrency-safe cache that keeps positive and negative
// results for different durations.
type Cache[K comparable, V any] struct {
	mu      sync.RWMutex
	entries map[K]entry[V]
	ttlHit  time.Duration
	ttlMiss time.Duration
}

// entry represents a cache entry with a value and an expiration time.
// V specifies the type of the value stored within the entry.
type entry[V any] struct {
	value  V
	expiry time.Time
}

// NewCache creates and returns a new instance of Cache with specified TTL durations for hits and misses.
func NewCache[K comparable, V any](ttlHit, ttlMiss time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		entries: make(map[K]entry[V]),
		ttlHit:  ttlHit,
		ttlMiss: ttlMiss,
	}
}

// Get retrieves the value associated with the given key from the cache, returning a boolean indicating existence.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok || !time.Now().Before(e.expiry) {
		var zero V
		return zero, false
	}
	return e.value, true
}

// Put stores a value in the cache with a key and sets an expiration based on whether the value is considered
// found or not.
func (c *Cache[K, V]) Put(key K, value V, found bool) {
	ttl := c.ttlHit
	if !found {
		ttl = c.ttlMiss
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry[V]{value: value, expiry: time.Now().Add(ttl)}
}

// Fetch retrieves a value associated with the key or loads it if not found in the cache, handling cache
// behavior accordingly.
func (c *Cache[K, V]) Fetch(ctx context.Context, key K, load func(context.Context) (V, error),
	found func(V) bool, onHit func(*V),
) (V, error) {
	if v, ok := c.Get(key); ok {
		onHit(&v)
		return v, nil
	}

	v, err := load(ctx)
	if err != nil {
		return v, err
	}

	c.Put(key, v, found(v))
	return v, nil
}
