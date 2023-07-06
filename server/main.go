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
	addr := flag.String("addr", ":8080", "The server address.")
	flag.Parse()

	log.Printf("Server running on address %s\n", *addr)
	return http.ListenAndServe(*addr, nil)
}
