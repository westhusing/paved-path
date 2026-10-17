// Package server gives every HTTP service in the fleet one shared lifecycle:
// listen, then drain on SIGINT/SIGTERM with a bounded timeout. Services stay
// small mains; the lifecycle lives here and is tested once, not per service.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run serves h on the PORT env var (default 8080, so one image works behind
// any host-port mapping) and shuts down gracefully on SIGINT/SIGTERM.
// It returns the listener's error if the server fails before any signal, or
// the shutdown error if draining times out.
func Run(log *slog.Logger, name string, h http.Handler) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second, // slowloris: bound the request headers
		ReadTimeout:       15 * time.Second, // whole read (headers + body)
		WriteTimeout:      20 * time.Second, // whole write; kept > ReadTimeout so a
		                   // slow read can't eat into the response budget
		IdleTimeout:       60 * time.Second, // keep-alive connections
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "service", name, "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop) // reset the handler when Run returns

	var sig os.Signal
	select {
	case err := <-errCh:
		return err // failed to listen; nothing to drain
	case sig = <-stop:
	}

	log.Info("shutting down", "service", name, "signal", sig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	log.Info("stopped", "service", name)
	return nil
}
