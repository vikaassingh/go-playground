package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type request struct {
	done chan struct{}
}

type LeakyBucket struct {
	queue    chan request
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewLeakyBucket(capacity, ratePerSec int) *LeakyBucket {
	if ratePerSec <= 0 {
		panic("reqPerCec must be > 0")
	}

	ctx, cancel := context.WithCancel(context.Background())

	lb := &LeakyBucket{
		queue:    make(chan request, capacity),
		interval: time.Second / time.Duration(ratePerSec),
		ctx:      ctx,
		cancel:   cancel,
	}

	lb.wg.Add(1)
	go lb.worker()

	return lb
}

func (lb *LeakyBucket) worker() {
	defer lb.wg.Done()

	ticker := time.NewTicker(lb.interval)
	defer ticker.Stop()

	for {
		select {
		case <-lb.ctx.Done():
			for {
				select {
				case req := <-lb.queue:
					// drain remaining requests gracefully
					close(req.done)
				default:
					return
				}
			}
		case <-ticker.C:
			select {
			case req := <-lb.queue:
				// processing request
				close(req.done)
			default:
				//nothing to process
			}
		}
	}
}

func (lb *LeakyBucket) Allow() error {
	req := request{done: make(chan struct{})}

	select {
	case lb.queue <- req:
		select {
		case <-req.done:
			return nil
		case <-lb.ctx.Done():
			return lb.ctx.Err()
		}
	default:
		return errors.New("rate limit exceeded")
	}
}

func (lb *LeakyBucket) Close() {
	lb.cancel()
	lb.wg.Wait()
}

func main() {
	lb := NewLeakyBucket(10, 5)
	defer lb.Close()

	var wg sync.WaitGroup

	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := lb.Allow()
			if err != nil {
				fmt.Println("Request", i, "rejected:", err)
				return
			}

			fmt.Println("Request", i, "accepted")
		}(i)
	}
	wg.Wait()
}
