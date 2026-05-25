package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Counter struct {
	count           int64
	expiryTimestamp time.Time

	id       int
	mu       sync.Mutex
	duration time.Duration
}

func (c *Counter) Increment(
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	ticker := time.NewTicker(c.duration)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:

			c.mu.Lock()

			c.count++

			fmt.Printf(
				"\nIncrementing counter-%d: %d",
				c.id,
				c.count,
			)

			c.mu.Unlock()
		}
	}
}

func (c *Counter) Reset(
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:

			c.mu.Lock()

			if time.Now().After(c.expiryTimestamp) {

				c.count = 0

				fmt.Printf(
					"\nResetting counter-%d: %d",
					c.id,
					c.count,
				)

				// reset next expiry
				c.expiryTimestamp =
					time.Now().Add(30 * time.Second)
			}

			c.mu.Unlock()
		}
	}
}

func NewCounter(
	id int,
	expiryTime time.Time,
	duration time.Duration,
) *Counter {

	return &Counter{
		id:              id,
		duration:        duration,
		expiryTimestamp: expiryTime,
	}
}

func main() {

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Second,
	)
	defer cancel()

	totalCounter := 2

	for i := 1; i <= totalCounter; i++ {

		counter := NewCounter(
			i,
			time.Now().Add(time.Duration(20*i)*time.Second),
			time.Duration(i)*time.Second,
		)

		wg.Add(2)

		go counter.Increment(ctx, wg)
		go counter.Reset(ctx, wg)
	}

	wg.Wait()
}
