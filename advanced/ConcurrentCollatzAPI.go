package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Request supports multiple numbers
type req struct {
	Nums []int `json:"nums"`
}

// Response includes winner + sequence
type resp struct {
	Num      int   `json:"num"`
	Sequence []int `json:"sequence"`
}

// Concurrent-safe cache
var cache sync.Map

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/collatz", collatzHandler)

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}

func collatzHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	data := &req{}
	if err := json.Unmarshal(body, data); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resultChan := make(chan resp, 1)

	var wg sync.WaitGroup

	for _, num := range data.Nums {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			// Check cache first
			if val, ok := cache.Load(n); ok {
				select {
				case resultChan <- resp{Num: n, Sequence: val.([]int)}:
					cancel()
				default:
				}
				return
			}

			seq := collatz(n)

			// Store in cache
			cache.Store(n, seq)

			select {
			case resultChan <- resp{Num: n, Sequence: seq}:
				cancel() // stop others
			case <-ctx.Done():
				return
			}
		}(num)
	}

	var result resp

	select {
	case result = <-resultChan:
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
		return
	}

	// Wait for cleanup (optional)
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	out, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func collatz(n int) []int {
	iter := []int{}
	for n != 1 {
		iter = append(iter, n)
		if n%2 == 0 {
			n = n / 2
		} else {
			n = 3*n + 1
		}
	}
	iter = append(iter, 1)
	time.Sleep(100 * time.Millisecond) // simulate work
	return iter
}
