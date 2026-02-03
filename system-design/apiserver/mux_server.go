package apiserver

/*
import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Run() {
	// func main() {
	// Router
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	// Apply middleware
	handler := Chain(
		mux,
		RecoveryMiddleware,
		LoggingMiddleware,
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	wg := &sync.WaitGroup{}

	// Start server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutdown signal received")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	} else {
		log.Println("server shutdown gracefully")
	}

	wg.Wait()
	log.Println("server exited")
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second) // simulate work
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
*/
