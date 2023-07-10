package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestDelete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     int
		path     string
		respBody string
	}{
		{name: "key exists", code: http.StatusOK, path: "/db?key=test"},
		{name: "key does not exist", code: http.StatusNotFound, path: "/?key=test1"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["test"] = "succeeded"
			s := testServer(&db)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			w := httptest.NewRecorder()
			s.serveHTTP(w, req)
			is.Equal(w.Code, tt.code)
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     int
		path     string
		respBody string
	}{
		{name: "test ok", code: http.StatusOK, path: "/db?key=test", respBody: "succeeded"},
		{name: "test key not found", code: http.StatusNotFound, path: "/db?key=not-there", respBody: "succeeded"},
		{name: "wrong value", code: http.StatusNotFound, path: "/?key=not"},
		{name: "no value", code: http.StatusNotFound, path: "/?key=", respBody: ""},
		{name: "wrong path", code: http.StatusNotFound, path: "/test/?key=not", respBody: ""},
		{name: "wrong path 2", code: http.StatusNotFound, path: "/test/test?key=not", respBody: ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["test"] = "succeeded"
			s := testServer(&db)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			s.serveHTTP(w, req)
			is.Equal(w.Code, tt.code)
			if tt.code == http.StatusOK {
				is.Equal(w.Body.String(), tt.respBody)
			}
		})
	}
}

func TestPost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		reqBody  string
		code     int
		response string
	}{
		{name: "simple", path: "/db?key=test", reqBody: "new-entry", code: http.StatusCreated, response: "succeeded"},
		{name: "weird body", path: "/db?key=test", reqBody: "ntest!@#$%^&*({ }+=)-/\\/test_;'\"", code: http.StatusCreated, response: "succeeded"},
		{name: "key exists", path: "/db?key=exists", reqBody: "", code: http.StatusConflict, response: "succeeded"},
		{name: "wrong path", path: "/test/test?key=not", reqBody: "new-entry", code: http.StatusNotFound, response: "succeeded"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["exists"] = "exists"

			s := testServer(&db)

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			s.serveHTTP(w, req)
			is.Equal(w.Code, tt.code)
		})
	}
}

func TestPut(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		reqBody  string
		code     int
		response string
	}{
		{name: "simple", path: "/db?key=test", reqBody: "new-entry", code: http.StatusCreated, response: "succeeded"},
		{name: "weird body", path: "/db?key=test", reqBody: "ntest!@#$%^&*({ }+=)-/\\/test_;'\"", code: http.StatusCreated, response: "succeeded"},
		{name: "empty body", path: "/db?key=test", reqBody: "", code: http.StatusCreated, response: "succeeded"},
		{name: "overwrite", path: "/db?key=exists", reqBody: "new-entry", code: http.StatusOK, response: "succeeded"},
		{name: "wrong path", path: "/test/test?key=not", reqBody: "new-entry", code: http.StatusNotFound, response: "succeeded"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["exists"] = "exists"

			s := testServer(&db)

			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			s.serveHTTP(w, req)
			is.Equal(w.Code, tt.code)
		})
	}
}

func TestParallel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		key     string
		reqBody string
		code    int
	}{
		{name: "simple", key: "test", reqBody: "new-entry", code: http.StatusCreated},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db := make(map[string]string)
			db["exists"] = "exists"
			s := testServer(&db)
			s.routes()
			is := is.New(t)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/db?key=%s", tt.key), strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			for i := 0; i < 5; i++ {
				go func() {
					s.mux.ServeHTTP(w, req)
					is.Equal(w.Code, tt.code)
				}()
			}
		})
	}
}

func testServer(db *map[string]string) *server {
	return &server{
		db: &database{
			db,
		},
		mux: http.NewServeMux(),
	}
}

func (s *server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.routes()
	s.mux.ServeHTTP(w, r)
}
