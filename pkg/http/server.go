package http

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	addr     string
	handlers *Handlers
	server   *http.Server
}

func NewServer(addr string, handlers *Handlers) *Server {
	return &Server{
		addr:     addr,
		handlers: handlers,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// REST API Endpoints
	mux.HandleFunc("/v1/limiter/check", s.handlers.Check)
	mux.HandleFunc("/v1/limiter/status", s.handlers.Status)
	mux.HandleFunc("/v1/limiter/reset", s.handlers.Reset)
	mux.HandleFunc("/v1/limiter/rules", s.handlers.Rules)
	mux.HandleFunc("/healthz", s.handlers.Healthz)
	mux.HandleFunc("/v1/stats", s.handlers.Stats)

	// Prometheus Metrics Endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Static Web Visualizer / Dashboard files
	webDir := "./web"
	if _, err := os.Stat(webDir); err == nil {
		fs := http.FileServer(http.Dir(webDir))
		mux.Handle("/web/", http.StripPrefix("/web/", fs))
		mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
				return
			}
			fs.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"service":"Distributed Rate Limiter","status":"healthy","endpoints":["/v1/limiter/check","/v1/limiter/status","/v1/limiter/rules","/healthz","/metrics"]}`))
		})
	}

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Close() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-RateLimit-Cost")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
