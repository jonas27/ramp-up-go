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
		{name: "works, key value exists", code: http.StatusOK, path: "/?key=test"},
		{name: "not working, key value exists", code: http.StatusNotFound, path: "/?key=test1"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["test"] = "succeeded"

			s := &server{
				db: &database{
					&db,
				},
			}
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
		{name: "works, test ok", code: http.StatusOK, path: "/?key=test", respBody: "succeeded"},
		{name: "not working, test ok", code: http.StatusNotFound, path: "/?key=not-there", respBody: "succeeded"},
		{name: "not working, wrong value", code: http.StatusNotFound, path: "/?key=not"},
		{name: "not working, no value", code: http.StatusNotFound, path: "/?key=", respBody: ""},
		{name: "not working, wrong path", code: http.StatusBadRequest, path: "/test/?key=not", respBody: ""},
		{name: "not working, wrong path 2", code: http.StatusBadRequest, path: "/test/test?key=not", respBody: ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			db := make(map[string]string)
			db["test"] = "succeeded"

			s := &server{
				db: &database{
					&db,
				},
			}
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

// TODO: for all tests: split path to path + key vars for better  testing
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
		{name: "works, simple", path: "/?key=test", reqBody: "new-entry", code: http.StatusCreated, response: "succeeded"},
		{name: "works, weird body", path: "/?key=test", reqBody: "ntest!@#$%^&*({ }+=)-/\\/test_;'\"", code: http.StatusCreated, response: "succeeded"},
		{name: "works, empty body", path: "/?key=test", reqBody: "", code: http.StatusCreated, response: "succeeded"},
		{name: "not working, wrong path", path: "/test/test?key=not", reqBody: "new-entry", code: http.StatusBadRequest, response: "succeeded"},
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

			s := &server{
				db: &database{
					&db,
				},
			}
			req := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			is.Equal(w.Code, tt.code)
		})
	}
}
