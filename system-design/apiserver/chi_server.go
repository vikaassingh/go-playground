package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ch "github.com/go-chi/chi/v5"
)

type Server struct {
	host    string
	port    string
	handler http.Handler
	server  *http.Server
}

func main() {
	svr := &Server{}
	svr.Run()
}

func (s *Server) Run() {
	s.host = "localhost"
	s.port = "8080"
	s.Start()
}

func (s *Server) Start() {
	s.server = s.NewServer()
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
		log.Printf("Server starting on %s:%s", s.host, s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Waiting for shutdown signal
	<-ctx.Done()
	log.Println("Shutdown signal received")
	s.Stop()

	wg.Wait()
	fmt.Println("Server exiting")
}

func (s *Server) NewServer() *http.Server {
	router := ch.NewRouter()
	s.Routes(router)
	return &http.Server{
		Addr:    s.host + ":" + s.port,
		Handler: router,
	}
}

func (s *Server) Stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Gracefully shutdown server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed:%+v", err)
	} else {
		log.Println("Server shutdown gracefully")
	}
}

func (s *Server) Routes(r *ch.Mux) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}
