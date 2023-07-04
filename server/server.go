package main

import (
	"log"
	"net/http"
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
	ok := checkPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := r.URL.Query().Get("key")
	value := s.db.get(key)
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

func checkPath(path string) bool {
	return path == "/"
}

func handleNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}
