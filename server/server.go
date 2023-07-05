package main

import (
	"io"
	"log"
	"net/http"
)

// server does not use a multiplexer, as we only care about http method, query params and request body
type server struct {
	db     *database
	server *http.Server
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := r.URL.Query().Get("key")
	switch r.Method {
	case http.MethodDelete:
		s.handleDelete(w, key)
	case http.MethodGet:
		s.handleGet(w, key)
	case http.MethodPut:
		s.handlePut(w, r, key)
	default:
		handleNotImplemented(w)
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

func handleNotImplemented(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotImplemented)
}
