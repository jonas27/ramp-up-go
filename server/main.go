package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	exitFail = 1
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
		server: &http.Server{
			Addr:              *port,
			ReadHeaderTimeout: 3 * time.Second,
		},
	}

	log.Printf("Server running on port %s\n", *port)
	return s.server.ListenAndServe()
}
