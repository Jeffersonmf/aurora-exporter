// Package run models pipeline executions and where their history is
// kept.
package run

import (
	"context"
	"errors"
	"time"
)

// Status of a run.
type Status string

// Run statuses.
const (
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

// Trigger says what started a run.
type Trigger string

// Run triggers.
const (
	TriggerSchedule Trigger = "schedule"
	TriggerAPI      Trigger = "api"
	TriggerCLI      Trigger = "cli"
)

// Run is one execution of one pipeline.
type Run struct {
	ID          string     `json:"id"`
	Pipeline    string     `json:"pipeline"`
	Trigger     Trigger    `json:"trigger"`
	Status      Status     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	DurationMS  int64      `json:"duration_ms"`
	RowsRead    int64      `json:"rows_read"`
	RowsWritten int64      `json:"rows_written"`
	Batches     int        `json:"batches"`
	Error       string     `json:"error,omitempty"`
}

// ErrNotFound is returned by Reader.Get for an unknown run ID.
var ErrNotFound = errors.New("run not found")

// Filter narrows Reader.List.
type Filter struct {
	// Pipeline, when set, keeps only that pipeline's runs.
	Pipeline string
	// Limit caps the result; zero means the store's default.
	Limit int
}

// Recorder persists runs. It is the only thing the service needs from
// a store, so the service depends on this narrow interface instead of
// a full repository (Interface Segregation).
type Recorder interface {
	Save(ctx context.Context, r Run) error
}

// Reader queries run history, newest first. The HTTP API depends on
// this and nothing else.
type Reader interface {
	Get(ctx context.Context, id string) (Run, error)
	List(ctx context.Context, f Filter) ([]Run, error)
}

// Store is a Recorder that can also be read back.
type Store interface {
	Recorder
	Reader
}
