package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*bucket
	maxTokens  float64
	refillRate float64
}

func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*bucket),
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     rl.maxTokens,
			lastRefill: now,
		}

		rl.buckets[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(rl.maxTokens, b.tokens+(elapsed*rl.refillRate))
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func main() {
	rl := NewRateLimiter(10, 5)
	var wg sync.WaitGroup

	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user := getUser()
			allowed := rl.Allow(user)
			fmt.Println(user, "Request", i, "Allowed", allowed)
			// if allowed {
			// }
		}()
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}

	wg.Wait()
}

func getUser() string {
	return fmt.Sprintf("user%v", rand.Intn(3))

}
