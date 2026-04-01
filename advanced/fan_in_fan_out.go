package main

import (
	"fmt"
	"sync"
)

// worker function
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for num := range jobs {
		fmt.Printf("Worker %d processing %d\n", id, num)
		results <- num * num
	}
}

func main() {
	numbers := []int{1, 2, 3, 4, 5, 6}

	jobs := make(chan int, len(numbers))
	results := make(chan int, len(numbers))

	var wg sync.WaitGroup

	// 🔹 Fan-out: start 3 workers
	numWorkers := 3
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	// send jobs
	for _, num := range numbers {
		jobs <- num
	}
	close(jobs)

	// 🔹 Close results after all workers done (fan-in coordination)
	go func() {
		wg.Wait()
		close(results)
	}()

	// 🔹 Fan-in: collect results
	for res := range results {
		fmt.Println("Result:", res)
	}
}
