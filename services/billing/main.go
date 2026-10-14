package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"example.com/paved-path/internal/billing"
	"example.com/paved-path/internal/health"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.Handle("/healthz", health.Handler())
	mux.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// Port comes from the environment so one image can be deployed behind
	// any host-port mapping; 8080 is the container default (see Dockerfile).
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	logger.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server", "err", err)
		os.Exit(1)
	}
}
