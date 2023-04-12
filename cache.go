package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExpired  = errors.New("expired")
)

type Value[T any] struct {
	Value T
	TTL   int64
}

func (v *Value[T]) IsExpired() bool {
	return v.TTL <= time.Now().Unix()
}

type MemoryCache[S comparable, T any] struct {
	TTL   time.Duration
	cache map[S]Value[T]
	mu    sync.RWMutex
}

func NewMemoryCache[S comparable, T any](ttl time.Duration) *MemoryCache[S, T] {
	return &MemoryCache[S, T]{
		TTL:   ttl,
		cache: make(map[S]Value[T]),
	}
}

func (c *MemoryCache[S, T]) Get(key S) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.cache[key]; ok {
		if v.IsExpired() {
			delete(c.cache, key)
			var noop T
			return noop, ErrExpired
		}
		return v.Value, nil
	}
	var noop T
	return noop, ErrNotFound
}

func (c *MemoryCache[S, T]) Set(key S, value T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = Value[T]{
		Value: value,
		TTL:   time.Now().Add(c.TTL).Unix(),
	}
	return nil
}

func (c *MemoryCache[S, T]) Delete(key S) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[key]; ok {
		delete(c.cache, key)
	}
}
func (c *MemoryCache[S, T]) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[S]Value[T])
	return nil
}

func (c *MemoryCache[S, T]) Count() (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache), nil
}

func (c *MemoryCache[S, T]) FlushExpired() {
	for k, v := range c.cache {
		if v.IsExpired() {
			c.Delete(k)
		}
	}
}

func (c *MemoryCache[S, T]) FlushExpiredLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.FlushExpired()
		case <-ctx.Done():
			return
		}
	}
}
