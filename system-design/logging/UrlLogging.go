package logging

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	urlList = []string{
		"http://example.com/page1",
		"http://example.com/page2",
		"http://example.com/page3",
	}
	responseCh = make(chan HttpResponse)
	wg         = &sync.WaitGroup{}
)

type HttpResponse struct {
	Url        string
	StatusCode int
	Message    string
}

func UrlLogging() {
	// wg.Add(1)
	go func() {
		// defer wg.Done()
		for res := range responseCh {
			log.Printf("URL: %s, Status Code: %d, Message: %s\n", res.Url, res.StatusCode, res.Message)
		}
	}()

	for _, url := range urlList {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			logUrl(url)
		}(url)
	}

	wg.Wait()
	close(responseCh)
}

func logUrl(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Error requesting URL %s: %v", url, err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error fetching URL %s: %v", url, err)
		return
	}
	defer res.Body.Close()
	// json.NewDecoder(res.Body).Decode(&res.Body)
	response := HttpResponse{
		Url:        url,
		StatusCode: res.StatusCode,
		Message:    res.Status,
	}
	responseCh <- response
}
