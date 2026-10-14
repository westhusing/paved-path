// Package jobs defines the job type and its processing logic.
// Pure function in, result out — the worker service is a thin loop around it.
package jobs

import "errors"

var ErrEmptyID = errors.New("job id must not be empty")

type Job struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
}

type Result struct {
	Status  string `json:"status"`
	Handled bool   `json:"handled"`
}

// Process validates and handles a job. Deterministic and side-effect free,
// so it is trivially testable and safe to retry.
func Process(j Job) (Result, error) {
	if j.ID == "" {
		return Result{}, ErrEmptyID
	}
	return Result{Status: "ok", Handled: true}, nil
}
