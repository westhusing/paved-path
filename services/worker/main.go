// Command worker consumes jobs and processes them one at a time.
// The demo keeps the queue in memory; the point is the shape: a small main
// around a pure, tested Process function that is safe to retry.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"example.com/paved-path/internal/jobs"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	queue := make(chan jobs.Job, 4)
	go func() {
		for i := 1; ; i++ {
			queue <- jobs.Job{ID: fmt.Sprintf("job-%04d", i), Payload: time.Now().UTC().Format(time.RFC3339)}
			time.Sleep(5 * time.Second)
		}
	}()

	for j := range queue {
		res, err := jobs.Process(j)
		if err != nil {
			logger.Warn("rejected", "job", j.ID, "err", err)
			continue
		}
		logger.Info("processed", "job", j.ID, "status", res.Status)
	}
}
