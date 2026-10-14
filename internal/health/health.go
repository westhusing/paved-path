// Package health provides the /healthz endpoint.
package health

import (
	"encoding/json"
	"net/http"
	"time"
)

var startedAt = time.Now()

// Handler returns a /healthz endpoint reporting liveness and uptime.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"uptime": time.Since(startedAt).String(),
		})
	})
}
