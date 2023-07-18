package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/goleak"
)

const (
	exists    = "exists"
	succeeded = "succeeded"
)

func TestDelete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		code int
		path string
		key  string
	}{
		{name: "key exists", code: http.StatusOK, path: "/db", key: "test"},
		{name: "key does not exist", code: http.StatusNotFound, path: "/db", key: "test1"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["test"] = succeeded
			s := testServer(db)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("%s?key=%s", tt.path, tt.key), nil)
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
		key      string
		respBody string
	}{
		{name: "test ok", code: http.StatusOK, path: "/db", key: "test", respBody: "succeeded"},
		{name: "test key not found", code: http.StatusNotFound, path: "/db", key: "not-there", respBody: "succeeded"},
		{name: "wrong value", code: http.StatusNotFound, path: "/", key: "not"},
		{name: "no value", code: http.StatusNotFound, path: "/", key: "", respBody: ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["test"] = succeeded
			s := testServer(db)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?key=%s", tt.path, tt.key), nil)
			w := httptest.NewRecorder()
			s.serveHTTP(w, req)
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
		name    string
		path    string
		key     string
		reqBody string
		code    int
	}{
		{name: "simple", path: "/db", key: "test", reqBody: "new-entry", code: http.StatusCreated},
		{
			name: "weird body", path: "/db", key: "test", reqBody: "ntest!@#$%^&*({ }+=)-/\\/test_;'\"",
			code: http.StatusCreated,
		},
		{name: "empty body", path: "/db", key: "test", reqBody: "", code: http.StatusCreated},
		{name: "overwrite", path: "/db", key: "exists", reqBody: "new-entry", code: http.StatusOK},
		{
			name: "key too long", path: "/db", key: "tooooooooooooooolong", reqBody: "new-entry",
			code: http.StatusRequestEntityTooLarge,
		},
		{name: "body too long", path: "/db", key: "short", code: http.StatusRequestEntityTooLarge, reqBody: `too
		ooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo
		ooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo long`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db[exists] = exists

			s := testServer(db)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("%s?key=%s", tt.path, tt.key), strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			s.serveHTTP(w, req)
			is.Equal(w.Code, tt.code)
		})
	}
}

func TestWrongPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		key      string
		respBody string
		code     int
	}{
		{name: "wrong path", code: http.StatusNotFound, path: "/test/", key: "not", respBody: ""},
		{name: "wrong path 2", code: http.StatusNotFound, path: "/test/test", key: "not", respBody: ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["test"] = "succeeded"
			s := testServer(db)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?key=%s", tt.path, tt.key), nil)
			w := httptest.NewRecorder()
			s.serveHTTP(w, req)
			is.Equal(w.Code, tt.code)
			if tt.code == http.StatusOK {
				is.Equal(w.Body.String(), tt.respBody)
			}
		})
	}
}

func TestDBFull(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		key     string
		reqBody string
		code    int
	}{
		{name: "db full", path: "/db", key: "normal", reqBody: "new-entry", code: http.StatusInsufficientStorage},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			for i := 0; i < 2001; i++ {
				db[strconv.Itoa(i)] = exists
			}
			s := testServer(db)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("%s?key=%s", tt.path, tt.key), strings.NewReader(tt.reqBody))
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
			s := testServer(db)
			s.routes()
			is := is.New(t)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/db?key=%s", tt.key), strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			for i := 0; i < 3; i++ {
				go func() {
					s.mux.ServeHTTP(w, req)
					is.Equal(w.Code, tt.code)
				}()
			}
		})
	}
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func testServer(db map[string]string) *server {
	httpRequestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	})
	return &server{
		db: &database{
			db: db,
		},
		mux:                  http.NewServeMux(),
		requestCounterMetric: httpRequestsTotal,
	}
}

func (s *server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.routes()
	s.mux.ServeHTTP(w, r)
}
