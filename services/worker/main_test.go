package main

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"example.com/paved-path/internal/jobs"
)

// captureHandler records log entries so a test can count how many jobs were
// processed without touching production code.
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func (h *captureHandler) count(msg string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := 0
	for _, r := range h.records {
		if r.Message == msg {
			n++
		}
	}
	return n
}

// The whole point of the shutdown path is that it must not deadlock: a worker
// killed mid-job used to lose in-flight work, and the fix (stop producing,
// wait for the producer's in-flight send to finish, close the queue, drain)
// is only correct if each step actually completes. This drives the real
// shutdown function (the same one main calls) and fails on a hang instead of
// blocking the test forever.
func TestShutdownDrainsWithoutDeadlock(t *testing.T) {
	h := &captureHandler{}
	logger := slog.New(h)

	queue := make(chan jobs.Job, 4)
	stopProducing := make(chan struct{})
	producerDone := make(chan struct{})
	go produce(queue, stopProducing, producerDone, time.Millisecond)

	// Let the producer fill the queue buffer.
	time.Sleep(20 * time.Millisecond)

	seqDone := make(chan struct{})
	go func() {
		defer close(seqDone)
		stop := make(chan struct{})
		runDone := make(chan struct{})
		go func() { run(queue, stop, logger); close(runDone) }()
		time.Sleep(5 * time.Millisecond) // let run start consuming
		close(stop)
		<-runDone

		shutdown(queue, stopProducing, producerDone, logger) // the real function
	}()

	select {
	case <-seqDone:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown sequence deadlocked")
	}

	if n := h.count("processed"); n == 0 {
		t.Fatal("expected at least one job to be processed, got none")
	}
}

// drain must process every queued job and return once the queue is closed.
func TestDrainProcessesAllQueuedJobs(t *testing.T) {
	h := &captureHandler{}
	logger := slog.New(h)

	queue := make(chan jobs.Job, 3)
	queue <- jobs.Job{ID: "a"}
	queue <- jobs.Job{ID: "b"}
	queue <- jobs.Job{ID: "c"}
	close(queue)

	drain(queue, logger) // must return (queue is closed)

	if n := h.count("processed"); n != 3 {
		t.Fatalf("drain processed %d jobs, want 3", n)
	}
}
