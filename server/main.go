package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	exitFail = 1
)

var (
	httpRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	})
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFail)
	}
}

func run(args []string) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		port = flags.String("port", ":8080", "The server port with colon")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	db := make(map[string]string)
	s := &server{
		db: &database{
			db: db,
		},
		mux: http.NewServeMux(),
	}

	ticker := time.NewTicker(100 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := s.db.persist()
				if err != nil {
					log.Println(err)
				}
			case <-quit:
				log.Println("stopping database persistent ticker")
				ticker.Stop()
				return
			}
		}
	}()
	defer close(quit)

	srv := &http.Server{
		Addr:              *port,
		ReadHeaderTimeout: 3 * time.Second,
	}
	srv.Handler = s.mux

	s.routes()

	log.Printf("Server running on port %s\n", *port)
	return srv.ListenAndServe()
}
