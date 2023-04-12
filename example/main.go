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
	v, _ := c.Get("key")
	fmt.Printf("value: %s\n", v)

	// Count
	count, _ := c.Count()
	fmt.Printf("count: %d\n", count)
	cancel()

	// Wait for expiration
	time.Sleep(time.Second)

	// Get after expiration
	v, err := c.Get("key")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// Count after expiration
	count, _ = c.Count()
	fmt.Printf("count: %d\n", count)

}
