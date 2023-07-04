package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	exitFail = 1
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFail)
	}
}

func run(args []string) error {
	port := flag.String("port", ":8080", "The server port with colon")
	flag.Parse()

	log.Printf("Server running on port %s\n", *port)
	return http.ListenAndServe(*port, nil)
}
