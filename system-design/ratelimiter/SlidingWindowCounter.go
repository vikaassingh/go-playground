package main

import (
	"fmt"
	"sync"
	"time"
)

type SlidingWindowLimiter struct {
	mu sync.Mutex

	limit          int
	windowSize     time.Duration
	currentWindow  int
	previousWindow int
	windowStart    time.Time
}

func NewSlidingWindowLimiter(limit int, windowSize time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:       limit,
		windowSize:  windowSize,
		windowStart: time.Now(),
	}
}

func (s *SlidingWindowLimiter) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.windowStart)

	// Move to next window
	if elapsed >= s.windowSize {
		s.previousWindow = s.currentWindow
		s.currentWindow = 0
		s.windowStart = now
		elapsed = 0
	}

	// Calculate overlap ratio
	overlapRatio := float64(s.windowSize-elapsed) / float64(s.windowSize)

	// Effective count
	effectiveCount := float64(s.previousWindow)*overlapRatio +
		float64(s.currentWindow)

	if effectiveCount >= float64(s.limit) {
		return false
	}

	s.currentWindow++
	return true
}

func main() {
	limiter := NewSlidingWindowLimiter(5, 10*time.Second)

	for i := 1; i <= 10; i++ {
		if limiter.Allow() {
			fmt.Printf("Request %d allowed\n", i)
		} else {
			fmt.Printf("Request %d blocked\n", i)
		}

		time.Sleep(1 * time.Second)
	}
}
