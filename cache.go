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
	Value *T
	TTL   int64
}

func (v *Value[T]) IsExpired() bool {
	return v.TTL <= time.Now().Unix()
}

type MemoryCache[T any] struct {
	TTL   time.Duration
	cache map[string]Value[T]
	mu    sync.RWMutex
}

func NewMemoryCache[T any](ttl time.Duration) *MemoryCache[T] {
	return &MemoryCache[T]{
		TTL:   ttl,
		cache: make(map[string]Value[T]),
	}
}

func (c *MemoryCache[T]) Get(key string) (*T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.cache[key]; ok {
		if v.IsExpired() {
			delete(c.cache, key)
			return nil, ErrExpired
		}
		return v.Value, nil
	}
	return nil, ErrNotFound
}

func (c *MemoryCache[T]) Set(key string, value T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = Value[T]{
		Value: &value,
		TTL:   time.Now().Add(c.TTL).Unix(),
	}
	return nil
}

func (c *MemoryCache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[key]; ok {
		delete(c.cache, key)
	}
}
func (c *MemoryCache[T]) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]Value[T])
	return nil
}

func (c *MemoryCache[T]) Count() (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache), nil
}

func (c *MemoryCache[T]) FlushExpired() {
	for k, v := range c.cache {
		if v.IsExpired() {
			c.Delete(k)
		}
	}
}

func (c *MemoryCache[T]) FlushExpiredLoop(ctx context.Context, interval time.Duration) {
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
