package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type server struct {
	db                   *database
	mux                  *http.ServeMux
	requestCounterMetric prometheus.Counter
}

func (s *server) routes() {
	s.mux.HandleFunc("/db", s.metricsMiddleware(s.requestLoggerMiddleware(s.handleDB())))
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

// according to rfc guidelines PUT should create or replace resources
// https://www.rfc-editor.org/rfc/rfc2616#section-9.6
func (s *server) handlePut(w http.ResponseWriter, r *http.Request, key string) {
	_, ok := s.db.get(key)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	err = s.db.put(key, string(body))
	var keyErr *KeyError
	var valueErr *ValueError
	var dbErr *DatabaseError
	switch {
	case errors.As(err, &keyErr):
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Println(err)
		}
		return
	case errors.As(err, &valueErr):
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Println(err)
		}
		return
	case errors.As(err, &dbErr):
		w.WriteHeader(http.StatusInsufficientStorage)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Println(err)
		}
		return
	}
	if !ok {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) metricsMiddleware(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.requestCounterMetric.Inc()
		hf(w, r)
	}
}

func (s *server) requestLoggerMiddleware(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		path := r.URL.Path
		key := r.URL.Query().Get("key")
		start := time.Now()
		defer log.Printf("a request called with method: %s, path: %s, key: %s, and took %d nanoseconds",
			method, path, key, time.Since(start))
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
	r.MustRegister(s.requestCounterMetric)
	s.mux.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{})) //nolint:exhaustruct
}
