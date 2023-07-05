package main

import (
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
			s.ServeHTTP(w, req)
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
			s.ServeHTTP(w, req)
			is.Equal(w.Code, tt.code)
			if tt.code == http.StatusOK {
				is.Equal(w.Body.String(), tt.respBody)
			}
		})
	}
}

func TestPut(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		path         string
		reqBody      string
		overwriteKey bool
		code         int
		response     string
	}{
		{name: "simple", path: "/db?key=test", reqBody: "new-entry", code: http.StatusCreated, response: "succeeded"},
		{name: "weird body", path: "/db?key=test", reqBody: "ntest!@#$%^&*({ }+=)-/\\/test_;'\"", code: http.StatusCreated, response: "succeeded"},
		{name: "empty body", path: "/db?key=test", reqBody: "", code: http.StatusCreated, response: "succeeded"},
		{name: "wrong path", path: "/test/test?key=not", reqBody: "new-entry", code: http.StatusNotFound, response: "succeeded"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			if tt.overwriteKey {
				db["test"] = tt.reqBody
			}

			s := testServer(&db)

			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			is.Equal(w.Code, tt.code)
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
