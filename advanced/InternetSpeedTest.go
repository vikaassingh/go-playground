package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/speed-test", speedTestHandler).Methods("GET")

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func speedTestHandler(w http.ResponseWriter, r *http.Request) {
	url := "http://ipv4.download.thinkbroadband.com/50MB.zip"

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read entire response to measure actual download
	bytesDownloaded, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	duration := time.Since(start).Seconds()

	// Speed in Mbps
	speedMbps := (float64(bytesDownloaded) * 8) / (duration * 1024 * 1024)

	result := fmt.Sprintf(
		"Downloaded: %.2f MB\nTime: %.2f sec\nSpeed: %.2f Mbps\n",
		float64(bytesDownloaded)/(1024*1024),
		duration,
		speedMbps,
	)

	w.Write([]byte(result))
}
