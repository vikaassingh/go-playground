package advanced

package main

import (
	"fmt"
	"sync"
)

var (
	oddTurn  = make(chan struct{}, 1)
	evenTurn = make(chan struct{}, 1)
)

func main() {
	ch := make(chan int)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i++ {
			ch <- i
		}
		close(ch)
	}()

	wg.Add(2)
	oddTurn <- struct{}{}
	go evenWorker(ch, wg)
	go oddWorker(ch, wg)

	wg.Wait()
	close(oddTurn)
	close(evenTurn)
}

func oddWorker(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		<-oddTurn
		num, ok := <-ch
		if !ok {
			evenTurn <- struct{}{}
			return
		}
		fmt.Println("O:", num)
		evenTurn <- struct{}{}
	}
}

func evenWorker(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		<-evenTurn
		num, ok := <-ch
		if !ok {
			oddTurn <- struct{}{}
			return
		}
		fmt.Println("E:", num)
		oddTurn <- struct{}{}
	}
}
