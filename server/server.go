package main

import (
	"io"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type server struct {
	db  *database
	mux *http.ServeMux
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.routes()
	s.mux.ServeHTTP(w, r)
}

func (s *server) routes() {
	s.mux.HandleFunc("/db", s.metricsMiddleware(s.handleDB()))
	s.registerMetrics()

	s.mux.HandleFunc("/*", s.handleBadPath())
}

func (s *server) handleDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		switch r.Method {
		case http.MethodDelete:
			s.handleDelete(w, key)
		case http.MethodGet:
			s.handleGet(w, key)
		case http.MethodPost:
			s.handlePost(w, r, key)
		case http.MethodPut:
			s.handlePut(w, r, key)
		default:
			s.handleNotImplemented(w)
		}
	}
}

func (s *server) handleDelete(w http.ResponseWriter, key string) {
	_, ok := s.db.get(key)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.db.delete(key)
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleGet(w http.ResponseWriter, key string) {
	value, ok := s.db.get(key)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_, err := w.Write([]byte(value))
	if err != nil {
		log.Println("Error writing response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// according to rfc guidlines POST should only create but not replace resources
// https://www.rfc-editor.org/rfc/rfc2616#section-9.5
func (s *server) handlePost(w http.ResponseWriter, r *http.Request, key string) {
	_, ok := s.db.get(key)
	if ok {
		w.WriteHeader(http.StatusConflict)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
	}
	s.db.put(key, string(body))
	w.WriteHeader(http.StatusCreated)
}

// according to rfc guidlines PUT should create or replace resources
// https://www.rfc-editor.org/rfc/rfc2616#section-9.6
func (s *server) handlePut(w http.ResponseWriter, r *http.Request, key string) {
	_, ok := s.db.get(key)
	if !ok {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
	}
	s.db.put(key, string(body))
}

func (s *server) metricsMiddleware(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpRequestsTotal.Inc()
		hf(w, r)
	}
}

func (s *server) handleNotImplemented(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func (s *server) handleBadPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *server) registerMetrics() {
	log.Println("registering metrics")
	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestsTotal)
	s.mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
}
