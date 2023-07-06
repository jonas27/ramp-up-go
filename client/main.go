package main

import (
	"os"

	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	logger.Error("Testing")
}
