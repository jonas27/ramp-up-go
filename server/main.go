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
	addr := flag.String("addr", ":8080", "The server address")
	flag.Parse()

	db := make(map[string]string)
	s := &server{
		db: &database{
			&db,
		},
		server: &http.Server{
			Addr:              *addr,
			ReadHeaderTimeout: 3 * time.Second,
		},
	}
 
	log.Printf("Server running on addr %s", *addr)
	return s.server.ListenAndServe()
}
