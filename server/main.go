package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
)

const (
	exitFail             = 1
	serverTimeoutSeconds = 3
	tickerSeconds        = 100
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	if err := run(os.Args, logger); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFail)
	}
}

func run(args []string, log *slog.Logger) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	addr := flags.String("addr", ":8080", "The server addr with colon")
	if err := flags.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	db := make(map[string]string)
	httpRequestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	})
	s := &server{
		log: log,
		db: &database{
			db: db,
		},
		mux:                  http.NewServeMux(),
		requestCounterMetric: httpRequestsTotal,
	}

	ticker := time.NewTicker(tickerSeconds * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := s.db.persist()
				if err != nil {
					log.Info(err.Error())
				}
			case <-quit:
				log.Info("stopping database persistent ticker")
				ticker.Stop()
				return
			}
		}
	}()
	defer func() {
		close(quit)
		time.Sleep(1 * time.Second)
	}()

	srv := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: serverTimeoutSeconds * time.Second,
	}
	srv.Handler = s.mux

	s.routes()

	log.Info("Server running", "address", *addr)
	return fmt.Errorf("the server failed with error: %w", srv.ListenAndServe())
}
