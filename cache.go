package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Value[T any] struct {
	Value *T
	TTL   int64
}

func (v *Value[T]) IsExpired() bool {
	return v.TTL < time.Now().Unix()
}

type Cache[T any] interface {
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, value T) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Count(ctx context.Context) (int, error)
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

func (c *MemoryCache[T]) Get(ctx context.Context, key string) (*T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.cache[key]; ok {
		if v.IsExpired() {
			return v.Value, nil
		}
		delete(c.cache, key)
	}
	return nil, errors.New("not found")
}

func (c *MemoryCache[T]) Set(ctx context.Context, key string, value T) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = Value[T]{
		Value: &value,
		TTL:   time.Now().Add(c.TTL).Unix(),
	}
	return nil
}

func (c *MemoryCache[T]) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
	return nil
}
func (c *MemoryCache[T]) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]Value[T])
	return nil
}

func (c *MemoryCache[T]) Count(ctx context.Context) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache), nil
}

func (c *MemoryCache[T]) FlushExpired(ctx context.Context) {
	for k, v := range c.cache {
		if v.IsExpired() {
			c.Delete(ctx, k)
		}
	}
}

func (c *MemoryCache[T]) FlushExpiredLoop(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.FlushExpired(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
