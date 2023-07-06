package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/exp/slog"
)

const (
	exitFail = 1
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	if err := run(os.Args, logger); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFail)
	}
}

type client struct {
	log *slog.Logger
}

func run(args []string, log *slog.Logger) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		host   = flags.String("host", "http://localhost:8080", "The host to send the request")
		method = flags.String("m", "", "The http method to be used")
		key    = flags.String("key", "", "The key of the request")
		value  = flags.String("value", "", "The value to be set for a key")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *key == "" {
		return fmt.Errorf("using any method without a key is not valid")
	}
	params := url.Values{}
	params.Set("key", *key)
	dbURL := fmt.Sprintf("%s/db?%s", *host, params.Encode())
	c := client{log: log}
	switch *method {
	case "delete":
		if *value != "" {
			return fmt.Errorf("using 'delete' method with value is not possible")
		}
		out, err := c.delete(dbURL)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	default:
		return fmt.Errorf("use either 'delete', 'get' or 'put' method")
	}
}

func (c *client) delete(url string) (string, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err = checkRespOK(resp); err != nil {
		c.log.Info(strconv.Itoa(resp.StatusCode))
		return "", err
	}
	return "deleted", nil
}

func checkRespOK(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("the request returned with http code: %d", resp.StatusCode)
	}
}
