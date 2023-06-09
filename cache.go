package cache

import (
	"context"
	"sync"
	"time"
)

type volatile[T any] struct {
	Value T
	TTL   int64
}

func (v *volatile[T]) IsExpired() bool {
	return v.TTL <= time.Now().Unix()
}

type MemoryCache[S comparable, T any] struct {
	TTL   time.Duration
	cache map[S]volatile[T]
	mu    sync.RWMutex
}

func NewMemoryCache[S comparable, T any](ttl time.Duration) *MemoryCache[S, T] {
	return &MemoryCache[S, T]{
		TTL:   ttl,
		cache: make(map[S]volatile[T]),
	}
}

func (c *MemoryCache[S, T]) Get(key S) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.cache[key]; ok {
		if v.IsExpired() {
			var noop T
			return noop, false
		}
		return v.Value, true
	}
	var noop T
	return noop, false
}

func (c *MemoryCache[S, T]) Set(key S, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = volatile[T]{
		Value: value,
		TTL:   time.Now().Add(c.TTL).Unix(),
	}
}

func (c *MemoryCache[S, T]) Delete(key S) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[key]; ok {
		delete(c.cache, key)
	}
}
func (c *MemoryCache[S, T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[S]volatile[T])
}

func (c *MemoryCache[S, T]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

func (c *MemoryCache[S, T]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.cache {
		if v.IsExpired() {
			delete(c.cache, k)
		}
	}
}

func (c *MemoryCache[S, T]) FlushInterval(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.Flush()
		case <-ctx.Done():
			return
		}
	}
}
