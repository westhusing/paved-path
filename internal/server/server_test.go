package server

import (
	"log/slog"
	"net/http"
	"testing"
)

// A bad PORT must fail fast with the listener error instead of blocking on a
// signal: Run returns it, it is not swallowed.
func TestRunReturnsListenError(t *testing.T) {
	t.Setenv("PORT", "99999") // out of range for a TCP port
	if err := Run(slog.Default(), "test", http.NewServeMux()); err == nil {
		t.Fatal("want an error for an invalid port, got nil")
	}
}
