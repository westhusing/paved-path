// Command gateway is the edge service: health and routing.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"example.com/paved-path/internal/health"
	"example.com/paved-path/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.Handle("/healthz", health.Handler())

	if err := server.Run(logger, "gateway", mux); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
