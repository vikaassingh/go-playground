package main

import (
	"fmt"
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	mu             sync.Mutex
	limit          int
	windowSize     time.Duration
	windowStart    time.Time
	currentWindow  int
	previousWindow int
}

func NewSlidingWindowCounter(limit int, windowSize time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		limit:       limit,
		windowSize:  windowSize,
		windowStart: time.Now(),
	}
}

func (s *SlidingWindowCounter) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.windowStart)

	if elapsed >= s.windowSize {
		if elapsed >= 2*s.windowSize {
			s.previousWindow = 0
		} else {
			s.previousWindow = s.currentWindow
		}
		s.currentWindow = 0
		s.windowStart = s.windowStart.Add(s.windowSize)
		elapsed = 0
	}

	overlapRatio := float64(s.windowSize-elapsed) / float64(s.windowSize)
	effectiveCount := float64(s.currentWindow) + float64(s.previousWindow)*overlapRatio

	if effectiveCount >= float64(s.limit) {
		return false
	}

	s.currentWindow++
	return true
}

func main() {
	rl := NewSlidingWindowCounter(5, 10*time.Second)
	for i := 1; i <= 10; i++ {
		if rl.Allow() {
			fmt.Printf("Request %d allowed \n", i)
		} else {
			fmt.Printf("Request %d blocked\n", i)
		}

		time.Sleep(1 * time.Second)
	}
}
