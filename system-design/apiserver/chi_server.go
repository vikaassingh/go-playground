package main

import "fmt"

type Server struct {
	host string
	port string
	server *http.Server
	isReady atomic.Bool
	isDraining atomic.Bool
	activeRequest atomic.Int64
	drainTimeout time.Duration
	baseContext context.Context
	cancel context.CancelFunc
}

func main() {
	baseCtx, cancel := context.WithCancel(context.Background())
	svr := &Server{
		host : "localhost",
		port : "8080",
		cancel: cancel,
		baseContext : baseCtx,
		drainTimeout : time.Secons * 5,
	}
	svr.Run()
}

func (s *Server) Run() {
	s.SetupServer()
	s.Start()
}

func (s *Server) SetupServer() {
	routes := chi.NewRouter()
	routes.Use(s.requestMiddleware)
	s.RegisterRoutes(routes)
	s.server = &http.Server{
		Addr : s.host + ":" + s.port,
		Handler : routes,
		BaseContext: func(net.Listener) context.Context {
            return s.baseContext
        },
	}
	s.isReady.Store(true)
}

func (s *Server) requestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.isDraining.Load() {
			http.Error(w, "Server draining", http.StatusServiceUnavailable)
			return
		}
		
		s.activeRequest.Add(1)
		defer s.activeRequest.Add(-1)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) RegisterRoutes(r chi.Router) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("OK"))
	})
	
	r.Get("/ready", s.readynessHandler)
}

func (s *Server) readynessHandler(w http.ResponseWriter, r *http.Request) {
	if !s.isReady.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func (s *Server) Start() {
	intSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func (){
		defer wg.Done()
		log.Printf("Server started listening on %v:%v", s.host, s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ListenAndServe error:%v", err)
		} else {
			log.Printf("Server listening on %s:%s", s.host, s.port)
		}
	}()
	
	// Waiting for stop signal
	<-intSignal.Done()
	log.Println("Shutdown signal received")
	
	s.GracefullyShutdown()
	
	wg.Wait()
	log.Println("Exiting server")
}

func (s *Server) GracefullyShutdown() {
	log.Println("Marking not ready")
	s.isReady.Store(false)
	
	log.Println("Entering draining mod")
	s.isDraining.Store(true)
	
	time.Sleep(5 * time.Second)
	
	s.server.SetKeepAlivesEnabled(false)
	
	waitStart := time.Now()
	for {
		active := s.activeRequest.Load()
		if active == 0 {
			break
		}
		
		log.Printf("Waiting for %d active requests...", active)
		time.Sleep(200 * time.Millisecond)
	}
	log.Printf("All active request completed in %v", time.Since(waitStart))
	
	s.Shutdown()
}

func (s *Server) Shutdown() {	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.drainTimeout)
	defer cancel()
	s.cancel()
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Gracefully shutdown failed: %v", err)
	} else {
		log.Printf("Gracefully shutdown successfully")
	}
}
