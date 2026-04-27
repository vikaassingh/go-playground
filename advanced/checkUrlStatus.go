package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Error      error
}

func checkURL(ctx context.Context, client *http.Client, url string) Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{URL: url, Error: err}
	}

	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return Result{
			URL:     url,
			Latency: latency,
			Error:   err,
		}
	}
	defer resp.Body.Close()

	return Result{
		URL:        url,
		StatusCode: resp.StatusCode,
		Latency:    latency,
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- Result, client *http.Client) {
	defer wg.Done()
	for url := range jobs {
		results <- checkURL(ctx, client, url)
	}
}

func main() {
	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://stackoverflow.com",
		"https://example.com",
		"https://golang.org",
		"https://amazon.com",
		"https://facebook.com",
		"https://twitter.com",
		"https://linkedin.com",
		"https://netflix.com",
	}

	// Timeout per request
	timeout := 3 * time.Second

	client := &http.Client{
		Timeout: timeout,
	}

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))

	// Limit concurrency (important for production)
	numWorkers := 5
	var wg sync.WaitGroup

	ctx := context.Background()

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobs, results, client)
	}

	// Send jobs
	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	// Wait for workers
	wg.Wait()
	close(results)

	// Print results
	for res := range results {
		if res.Error != nil {
			fmt.Printf("❌ %s -> ERROR: %v (Latency: %v)\n", res.URL, res.Error, res.Latency)
		} else {
			fmt.Printf("✅ %s -> Status: %d (Latency: %v)\n", res.URL, res.StatusCode, res.Latency)
		}
	}
}
