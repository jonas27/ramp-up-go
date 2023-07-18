package main

import (
	"errors"
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
				httpmock.NewStringResponder(http.StatusOK, ``))

			httpmock.RegisterResponder(http.MethodDelete, "http://test.com/db?key=not-there",
				httpmock.NewStringResponder(http.StatusBadRequest, ``))

			params := url.Values{}
			params.Set("key", tt.key)
			dbURL := fmt.Sprintf("%s/db?%s", "http://test.com", params.Encode())

			c := client{log: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
			out, err := c.delete(dbURL)
			if tt.isErr {
				reqErro := requestError{tt.code}
				is.Equal(reqErro.Error(), err.Error())
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
				httpmock.NewStringResponder(http.StatusOK, `test-value`))

			httpmock.RegisterResponder(http.MethodGet, "http://test.com/db?key=not-there",
				httpmock.NewStringResponder(http.StatusBadRequest, ``))

			params := url.Values{}
			params.Set("key", tt.key)
			dbURL := fmt.Sprintf("%s/db?%s", "http://test.com", params.Encode())

			c := client{log: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
			out, err := c.get(dbURL)
			if tt.isErr {
				is.Equal(checkRespOK(tt.code), err)
			} else {
				is.NoErr(err)
				is.Equal(out, tt.value)
			}
		})
	}
}

func TestPut(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		key   string
		isErr bool
		value string
	}{
		{name: "simple existing", code: http.StatusOK, key: "test", value: "test-value", isErr: false},
		{name: "simple new", code: http.StatusCreated, key: "test-new", value: "test-new-value", isErr: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			is := is.New(t)
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(http.MethodPut, "http://test.com/db?key=test",
				httpmock.NewStringResponder(http.StatusOK, ``))
			httpmock.RegisterResponder(http.MethodPut, "http://test.com/db?key=test-new",
				httpmock.NewStringResponder(http.StatusCreated, ``))

			params := url.Values{}
			params.Set("key", tt.key)
			dbURL := fmt.Sprintf("%s/db?%s", "http://test.com", params.Encode())

			c := client{log: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
			out, err := c.put(dbURL, tt.value)
			if tt.isErr {
				is.Equal(fmt.Errorf("the request returned with http code: %d", tt.code), err)
			} else {
				is.NoErr(err)
				if tt.code == http.StatusOK {
					is.Equal(out, "updated")
				} else {
					is.Equal(out, "created")
				}
			}
		})
	}
}

func TestPutLimits(t *testing.T) {
	tests := []struct {
		name  string
		code  int
		key   string
		value string
	}{
		{name: "key too long", key: "tooooooooooooooolong", value: "new-entry", code: http.StatusRequestEntityTooLarge},
		{name: "body too long", key: "largebody", code: http.StatusRequestEntityTooLarge, value: `too
		ooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo
		ooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo long`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(http.MethodPut, "http://test.com/db?key=tooooooooooooooolong",
				httpmock.NewStringResponder(http.StatusRequestEntityTooLarge, ``))
			httpmock.RegisterResponder(http.MethodPut, "http://test.com/db?key=largebody",
				httpmock.NewStringResponder(http.StatusRequestEntityTooLarge, ``))

			params := url.Values{}
			params.Set("key", tt.key)
			dbURL := fmt.Sprintf("%s/db?%s", "http://test.com", params.Encode())

			c := client{log: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
			_, err := c.put(dbURL, tt.value)
			var reqErr *requestError
			if errors.As(err, &reqErr) {
				is.Equal(reqErr.code, tt.code)
			}

		})
	}
}

func TestRunDelete(t *testing.T) {
	is := is.New(t)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodDelete, "http://test.com/db?key=test",
		httpmock.NewStringResponder(200, `test-value`))

	err := run([]string{"test", "-host", "http://test.com", "-m", "delete", "-key", "test"}, logger)
	is.NoErr(err)
}

func TestRunGet(t *testing.T) {
	is := is.New(t)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, "http://test.com/db?key=test",
		httpmock.NewStringResponder(200, ``))

	err := run([]string{"test", "-host", "http://test.com", "-m", "get", "-key", "test"}, logger)
	is.NoErr(err)
}

func TestRunPut(t *testing.T) {
	is := is.New(t)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPut, "http://test.com/db?key=test",
		httpmock.NewStringResponder(201, `new-value`))

	err := run([]string{"test", "-host", "http://test.com", "-m", "put", "-key", "test", "-value", "new-value"}, logger)
	is.NoErr(err)
}
