package advanced

import (
	"fmt"
	"sync"
)

func OddEvenSync() {
	ch := make(chan int)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go workerOdd(ch, wg)
	ch <- 1
	go workerEven(ch, wg)

	wg.Wait()
}

func workerOdd(ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for num := range ch {
		fmt.Println(num, " : From Odd")
		if num < 10 {
			ch <- num + 1
		}
	}
}

func workerEven(ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for num := range ch {
		fmt.Println(num, " : From Even")
		if num < 10 {
			ch <- num + 1
		}
		if num == 10 {
			close(ch)
			return
		}
	}
}
