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
	maxWorkers   = 5
	maxRetries   = 3
	totalJobs    = 20
	processDelay = time.Millisecond * 300
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
	for workerID := range maxWorkers {
		wg.Add(1)
		go worker(ctx, workerID, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for job := range totalJobs {
			select {
			case <-ctx.Done():
				return
			case jobs <- Job{ID: job}:
			}
		}
	}()

	go func() {
		defer close(results)
		wg.Wait()
	}()

	var successFullResults []int
	for res := range results {
		successFullResults = append(successFullResults, res)
	}

	fmt.Println("successfull processed jobs:", successFullResults)
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
				case <-ctx.Done():
					return
				case results <- job.ID:
				}
			}
		}
	}
}

func processWithRetry(ctx context.Context, job Job, workerID int) bool {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		time.Sleep(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			return false
		default:
		}

		err := process(job)
		if err == nil {
			var mu sync.Mutex
			mu.Lock()
			defer mu.Unlock()
			printJobStatus("successfully processed", workerID, job, attempt)
			return true
		}

		printJobStatus("failed", workerID, job, attempt)
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

func printJobStatus(status string, workerID int, job Job, attempt int) {
	fmt.Printf("Worker %d %s Job %d (attempt %d)\n", workerID, status, job.ID, attempt)
}
