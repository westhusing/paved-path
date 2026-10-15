// Command billing serves invoice quotes. Money is integer cents, tax rates
// are basis points; see internal/billing for the arithmetic.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"example.com/paved-path/internal/billing"
	"example.com/paved-path/internal/health"
	"example.com/paved-path/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := server.Run(logger, "billing", routes()); err != nil {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}

// routes builds the billing service's mux. It is a plain function so handler
// behavior (query parsing, error mapping) is testable without a live server.
func routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", health.Handler())
	mux.HandleFunc("/quote", handleQuote)
	return mux
}

func handleQuote(w http.ResponseWriter, r *http.Request) {
	cents, err1 := strconv.ParseInt(r.URL.Query().Get("cents"), 10, 64)
	bps, err2 := strconv.Atoi(r.URL.Query().Get("bps"))
	if err1 != nil || err2 != nil {
		http.Error(w, "cents and bps are required", http.StatusBadRequest)
		return
	}
	q, err := billing.Compute(cents, bps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}
