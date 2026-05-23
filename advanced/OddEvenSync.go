package main

import (
	"context"
	"fmt"
)

func main() {
	var start, end int = 1, 10
	oddCh := make(chan int)
	evenCh := make(chan int)
	receiver := make(chan int)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(evenCh)
		for num := range oddCh {
			receiver <- num
			if num >= end {
				cancel()
				return
			}
			evenCh <- num + 1
		}
	}()

	go func() {
		defer close(oddCh)
		for num := range evenCh {
			receiver <- num
			if num >= end {
				cancel()
				return
			}
			oddCh <- num + 1
		}
	}()

	go func() {
		oddCh <- start
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case num := <-receiver:
			fmt.Println(num)
		}
	}
}
