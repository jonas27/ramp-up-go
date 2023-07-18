package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
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
	srv := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: serverTimeoutSeconds * time.Second,
	}
	srv.Handler = s.mux
	s.routes()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	errWg, errCtx := errgroup.WithContext(ctx)

	ticker := time.NewTicker(tickerSeconds * time.Second)
	errWg.Go(func() error {
		for {
			select {
			case <-ticker.C:
				err := s.db.persist()
				if err != nil {
					return err
				}
			case <-errCtx.Done():
				log.Info("stopping database and persist to disk")
				ticker.Stop()
				err := s.db.persist()
				if err != nil {
					return fmt.Errorf("could not persist db to disk: %w", err)
				}
				stop()
				return nil
			}
		}
	})

	errWg.Go(func() error {
		log.Info("Server running", "address", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("the server failed with error: %w", err)
		}
		return nil
	})

	errWg.Go(func() error {
		<-errCtx.Done()
		// https://gist.github.com/s8508235/bc248d046d5001d5cae46cc39066cdf5?permalink_comment_id=4360249#gistcomment-4360249
		return srv.Shutdown(context.Background())
	})

	err := errWg.Wait()
	if err == context.Canceled || err == nil {
		fmt.Println("gracefully quit server")
		return nil
	}
	return err
}
