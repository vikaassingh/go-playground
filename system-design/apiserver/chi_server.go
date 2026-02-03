package apiserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	host    string
	port    string
	stop    chan os.Signal
	handler http.Handler
}

func Run() {
	s := &Server{
		host:    "localhost",
		port:    "8080",
		stop:    make(chan os.Signal, 1),
		handler: nil, // Set your handler here
	}
	s.Start()
}
func (s *Server) Start() {
	router := s.NewRouter()
	svr := &http.Server{
		Addr:    s.host + ":" + s.port,
		Handler: router,
	}

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
		log.Printf("Server started on %s:%s", s.host, s.port)
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutdown signal received")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svr.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	} else {
		log.Println("Server Shutdown gracefully")
	}

	wg.Wait()
	log.Println("Server exiting")
}

func (s *Server) NewRouter() *chi.Mux {
	return chi.NewRouter()
}
