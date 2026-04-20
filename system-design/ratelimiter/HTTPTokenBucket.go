package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"strings"
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
}

func NewRateLimiter(maxTokens int, refillRate float64, ttl time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*bucket),
		maxTokens:  float64(maxTokens),
		refillRate: refillRate,
		ttl:        ttl,
	}

	go rl.cleanup()
	return rl
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

func (rl *RateLimiter) Allow(key string) (bool, float64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b := rl.getBucket(key)

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = math.Min(rl.maxTokens, b.tokens+elapsed*rl.refillRate)
	b.lastRefill = now

	if b.tokens >= 1 {
		b.tokens -= 1
		return true, b.tokens
	}

	return false, b.tokens
}

// Cleanup old buckets to prevent memory leak
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.ttl)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for k, b := range rl.buckets {
			if now.Sub(b.lastRefill) > rl.ttl {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// Extract client IP safely
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	return ip
}

// Middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getIP(r)

		allowed, tokens := rl.Allow(key)
		if !allowed {
			retryAfter := int(math.Ceil(1 / rl.refillRate))

			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// Optional headers
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", tokens))

		next.ServeHTTP(w, r)
	})
}

func main() {
	rl := NewRateLimiter(10, 5, 5*time.Minute) // 10 burst, 5 req/sec

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	handler := rl.Middleware(mux)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
