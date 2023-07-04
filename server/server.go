package main

import (
	"log"
	"net/http"
	"strings"
)

type server struct {
	db     *database
	server *http.Server
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	default:
		handleNotFound(w)
	}
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	paths := strings.Split(path, "/")
	if len(paths) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value := s.db.get(paths[0])
	if value == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_, err := w.Write([]byte(value))
	if err != nil {
		log.Println("Error writing response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

func handleNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}
