package main

import (
	"log"
	"sync"
	"time"
)

type SlidingWindow struct {
	mu         sync.Mutex
	window     time.Duration
	limit      int
	timestamps []time.Time
}

func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutOff := now.Add(-sw.window)

	idx := 0
	for idx < len(sw.timestamps) && sw.timestamps[idx].Before(cutOff) {
		idx++
	}
	sw.timestamps = sw.timestamps[idx:]
	if len(sw.timestamps) >= sw.limit {
		return false
	}

	sw.timestamps = append(sw.timestamps, now)
	return true
}

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
)

func main() {
	sw := &SlidingWindow{
		limit:  5,
		window: 2 * time.Second,
	}

	for i := 1; i <= 20; i++ {
		if sw.Allow() {
			log.Println(Green, "Request", i, "allowed", Reset)
		} else {
			log.Println(Red, "Request", i, "rejected", Reset)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
