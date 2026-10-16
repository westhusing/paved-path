// Command worker consumes jobs and processes them one at a time.
// The demo keeps the queue in memory; the point is the shape: a small main
// around a pure, tested Process function that is safe to retry.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/paved-path/internal/jobs"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	queue := make(chan jobs.Job, 4)
	stopProducing := make(chan struct{})
	producerDone := make(chan struct{})
	go produce(queue, stopProducing, producerDone, 5*time.Second)

	// Graceful shutdown: on SIGINT/SIGTERM stop producing, wait for the
	// producer to finish its in-flight send, close the queue, then drain the
	// jobs already queued. Without this a container kill mid-job loses work
	// that was in flight; the HTTP services get the same treatment from
	// internal/server.Run.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	stop := make(chan struct{})
	go func() {
		s := <-sig
		logger.Info("shutting down", "signal", s)
		close(stop)
	}()

	run(queue, stop, logger) // consume until the signal arrives
	close(stopProducing)     // no new jobs
	<-producerDone          // producer finished its in-flight send
	close(queue)            // safe now: no more sends will happen
	drain(queue, logger)    // process what's left, then return
}

// produce enqueues a job every interval until stop is closed. It selects on
// the send so a shutdown signal can't be blocked behind a full queue, and it
// signals done when it returns.
func produce(queue chan<- jobs.Job, stop <-chan struct{}, done chan<- struct{}, interval time.Duration) {
	defer close(done)
	for i := 1; ; i++ {
		select {
		case <-stop:
			return
		case queue <- jobs.Job{ID: fmt.Sprintf("job-%04d", i), Payload: time.Now().UTC().Format(time.RFC3339)}:
		}
		select {
		case <-stop:
			return
		case <-time.After(interval):
		}
	}
}

// run consumes from queue until stop is closed, processing each job. It does
// not drain: the caller drains after the producer has stopped (see main).
func run(queue <-chan jobs.Job, stop <-chan struct{}, logger *slog.Logger) {
	for {
		select {
		case <-stop:
			return
		case j := <-queue:
			process(j, logger)
		}
	}
}

// drain processes every job already in the queue, then returns. It only
// terminates because main closed the queue after the producer stopped.
func drain(queue <-chan jobs.Job, logger *slog.Logger) {
	for j := range queue {
		process(j, logger)
	}
}

func process(j jobs.Job, logger *slog.Logger) {
	res, err := jobs.Process(j)
	if err != nil {
		logger.Warn("rejected", "job", j.ID, "err", err)
		return
	}
	logger.Info("processed", "job", j.ID, "status", res.Status)
}
