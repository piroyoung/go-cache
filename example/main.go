package main

import (
	"context"
	"fmt"
	"github.com/piroyoung/go-cache"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize
	c := cache.NewMemoryCache[string, string](time.Second)
	go c.FlushExpiredLoop(ctx, time.Minute)
	defer cancel()

	// Set
	c.Set("key", "value")

	// Get
	v, ok := c.Get("key")
	fmt.Printf("value: %s, ok: %t\n", v, ok)

	// Count
	count := c.Count()
	fmt.Printf("count: %d\n", count)
	cancel()

	// Wait for expiration
	time.Sleep(time.Second)

	// Get after expiration
	v, ok = c.Get("key")
	fmt.Printf("value: %s, ok: %t\n", v, ok)

	// Count after expiration
	count = c.Count()
	fmt.Printf("count: %d\n", count)

}
