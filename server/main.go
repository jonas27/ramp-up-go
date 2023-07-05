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
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFail)
	}
}

func run() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	port := flag.String("port", ":8080", "The server port with colon")
	flag.Parse()

	db := make(map[string]string)
	s := &server{
		db: &database{
			&db,
		},
		mux: http.NewServeMux(),
	}

	srv := &http.Server{
		Addr:              *port,
		ReadHeaderTimeout: 3 * time.Second,
	}
	srv.Handler = s.mux

	s.routes()

	log.Printf("Server running on port %s\n", *port)
	return srv.ListenAndServe()
}
