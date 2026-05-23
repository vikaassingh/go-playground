package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	cache sync.Map
	wg    = &sync.WaitGroup{}
)

func main() {
	nums := []int{1, 2, 3, 4, 5, 6}
	numCh := make(chan int)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case num := <-numCh:
				cache.Store(num, num*num)
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer close(numCh)
		for _, num := range nums {
			select {
			case <-ctx.Done():
				return
			case numCh <- num:
			}
		}
	}()

	wg.Wait()

	for _, num := range nums {
		if square, ok := cache.Load(num); ok {
			fmt.Println(num, ":", square)
		} else {
			fmt.Println(num, ": unprocessed due to timeout")
		}
	}
}
