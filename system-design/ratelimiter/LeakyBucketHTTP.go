package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type request struct {
	done chan struct{}
}

type bucket struct {
	queue      chan request
	lastActive time.Time
}

type LeakyRateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	capacity  int
	interval  time.Duration
	ttl       time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewLeakyRateLimiter(capacity int, ratePerSec int, ttl time.Duration) *LeakyRateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &LeakyRateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: capacity,
		interval: time.Second / time.Duration(ratePerSec),
		ttl:      ttl,
		ctx:      ctx,
		cancel:   cancel,
	}

	rl.wg.Add(1)
	go rl.cleanup()

	return rl
}

func (rl *LeakyRateLimiter) getBucket(key string) *bucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{
			queue:      make(chan request, rl.capacity),
			lastActive: time.Now(),
		}
		rl.buckets[key] = b

		rl.wg.Add(1)
		go rl.worker(b)
	}

	b.lastActive = time.Now()
	return b
}

func (rl *LeakyRateLimiter) worker(b *bucket) {
	defer rl.wg.Done()

	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			select {
			case req := <-b.queue:
				close(req.done)
			default:
			}
		}
	}
}

// Non-blocking (fast fail)
func (rl *LeakyRateLimiter) Allow(key string) bool {
	b := rl.getBucket(key)

	select {
	case b.queue <- request{done: make(chan struct{})}:
		return true
	default:
		return false
	}
}

// Blocking (backpressure)
func (rl *LeakyRateLimiter) Wait(ctx context.Context, key string) error {
	b := rl.getBucket(key)
	req := request{done: make(chan struct{})}

	select {
	case b.queue <- req:
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("rate limited")
	}

	select {
	case <-req.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Cleanup inactive buckets
func (rl *LeakyRateLimiter) cleanup() {
	defer rl.wg.Done()

	ticker := time.NewTicker(rl.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for k, b := range rl.buckets {
				if now.Sub(b.lastActive) > rl.ttl {
					delete(rl.buckets, k)
				}
			}
			rl.mu.Unlock()
		}
	}
}

func (rl *LeakyRateLimiter) Close() {
	rl.cancel()
	rl.wg.Wait()
}

// Extract IP
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}
	ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	return ip
}

// Middleware
func (rl *LeakyRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getIP(r)

		// choose one:
		// 1. fast fail:
		if !rl.Allow(key) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// 2. OR blocking:
		/*
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := rl.Wait(ctx, key); err != nil {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		*/

		next.ServeHTTP(w, r)
	})
}

func main() {
	rl := NewLeakyRateLimiter(
		10,              // queue size (burst)
		5,               // leak rate (req/sec)
		5*time.Minute,   // TTL cleanup
	)
	defer rl.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	handler := rl.Middleware(mux)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
