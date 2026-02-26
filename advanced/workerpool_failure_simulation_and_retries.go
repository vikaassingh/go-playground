package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

/*
Implement controlled concurrency using a fixed number of worker goroutines (worker pool pattern).
Each job should simulate processing and may randomly fail.
If a job fails, it must be retried up to a maximum of 3 attempts.
Use context.Context to support graceful shutdown of all running goroutines.
Ensure the program does not cause any goroutine leaks.
After all processing is complete, print the IDs of successfully processed jobs.
*/
var (
	maxWorker    = 5
	maxRetries   = 3
	totalJobs    = 20
	processDelay = 300 * time.Millisecond
)

type Job struct {
	ID int
}

func main() {
	rand.Seed(time.Now().UnixNano())
	jobs := make(chan Job)
	results := make(chan int)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 1; i <= maxWorker; i++ {
		wg.Add(1)
		go worker(ctx, i, jobs, results, &wg)
	}
	go func() {
		defer close(jobs)
		for i := 1; i <= totalJobs; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- Job{ID: i}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var successFullJobs []int
	for id := range results {
		successFullJobs = append(successFullJobs, id)
	}

	fmt.Println("successfull processed jobs:", successFullJobs)
}

func worker(ctx context.Context, workerID int, jobs <-chan Job, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			if processWithRetry(ctx, job, workerID) {
				select {
				case results <- job.ID:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func processWithRetry(ctx context.Context, job Job, workerID int) bool {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		err := process(job)
		if err == nil {
			fmt.Printf("Worker %d successfully processed Job %d (attempt %d)\n", workerID, job.ID, attempt)
			return true
		}

		fmt.Printf("Worker %d failed job %d (attempt %d)\n", workerID, job.ID, attempt)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("Job %d permanently failed after %d attempts\n", job.ID, maxRetries)
	return false
}

func process(job Job) error {
	time.Sleep(processDelay)
	if rand.Intn(100) < 30 {
		return fmt.Errorf("random failure")
	}

	return nil
}
