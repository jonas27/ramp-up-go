package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/matryer/is"
	"golang.org/x/exp/slog"
)

func TestDelete(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		key   string
		isErr bool
	}{
		{name: "simple", code: http.StatusOK, key: "test", isErr: false},
		{name: "key does not exist", code: http.StatusBadRequest, key: "not-there", isErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(http.MethodDelete, "http://test.com/db?key=test",
				httpmock.NewStringResponder(200, ``))

			httpmock.RegisterResponder(http.MethodDelete, "http://test.com/db?key=not-there",
				httpmock.NewStringResponder(400, ``))

			params := url.Values{}
			params.Set("key", tt.key)
			dbURL := fmt.Sprintf("%s/db?%s", "http://test.com", params.Encode())

			c := client{log: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
			out, err := c.delete(dbURL)
			if tt.isErr {
				is.Equal(fmt.Errorf("the request returned with http code: %d", tt.code), err)
			} else {
				is.NoErr(err)
				is.Equal(out, "deleted")
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		key   string
		isErr bool
		value string
	}{
		{name: "simple", code: http.StatusOK, key: "test", value: "test-value", isErr: false},
		{name: "key does not exist", code: http.StatusBadRequest, key: "not-there", isErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(http.MethodGet, "http://test.com/db?key=test",
				httpmock.NewStringResponder(200, `test-value`))

			httpmock.RegisterResponder(http.MethodGet, "http://test.com/db?key=not-there",
				httpmock.NewStringResponder(400, ``))

			params := url.Values{}
			params.Set("key", tt.key)
			dbURL := fmt.Sprintf("%s/db?%s", "http://test.com", params.Encode())

			c := client{log: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
			out, err := c.get(dbURL)
			if tt.isErr {
				// t.Fatal(checkRespOK(tt.code))
				is.Equal(checkRespOK(tt.code), err)
			} else {
				is.NoErr(err)
				is.Equal(out, tt.value)
			}
		})
	}
}
