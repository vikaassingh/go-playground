package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
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
	ttl        time.Duration
	ctx        context.Context
}

func NewRateLimiyter(maxTokens, refillRate float64, ttl time.Duration, ctx context.Context) *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*bucket),
		maxTokens:  maxTokens,
		refillRate: refillRate,
		ttl:        ttl,
		ctx:        ctx,
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()

			for k, b := range rl.buckets {
				if now.Sub(b.lastRefill) > rl.ttl {
					log.Println("Deleting key:", k)
					delete(rl.buckets, k)
				}
			}
			rl.mu.Unlock()
		case <-rl.ctx.Done():
			fmt.Println("cleanup stoped")
			return
		}
	}
}

func (rl *RateLimiter) Apply(key string) (bool, float64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b := rl.getBucket(key)
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	tokens := b.tokens + (elapsed * rl.refillRate)
	if rl.maxTokens < tokens {
		tokens = rl.maxTokens
	}
	b.tokens = tokens
	b.lastRefill = now
	if b.tokens > 0 {
		b.tokens--
		return true, b.tokens
	}

	return false, b.tokens
}

func (rl *RateLimiter) getBucket(key string) *bucket {
	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     rl.maxTokens,
			lastRefill: time.Now(),
		}

		rl.buckets[key] = b
	}

	return b
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		randUser := GetRandUser(1000)
		allowed, tokens := rl.Apply(randUser)
		if !allowed {
			retryAfter := int(math.Ceil(1 / rl.refillRate))

			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", tokens))
		next.ServeHTTP(w, r)
	})
}

func GetRandUser(randUser int) string {
	return fmt.Sprintf("x-User-ID-%v", rand.Intn(randUser))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	rl := NewRateLimiyter(10, 5, 30*time.Second, ctx)
	defer cancel()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	handler := rl.Middleware(mux)

	fmt.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
