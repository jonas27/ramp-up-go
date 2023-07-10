package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     int
		path     string
		response string
	}{
		{name: "test ok", code: http.StatusOK, path: "/test", response: "succeeded"},
		{name: "wrong value", code: http.StatusNotFound, path: "/not", response: ""},
		{name: "no value", code: http.StatusNotFound, path: "/", response: ""},
		{name: "wrong path", code: http.StatusBadRequest, path: "/test/", response: ""},
		{name: "wrong path 2", code: http.StatusBadRequest, path: "/test/test", response: ""},
	}
	for _, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
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
				is.Equal(w.Body.String(), tt.response)
			}
		})
	}
}
