// Package service is the application layer: it runs pipelines by
// name, guarantees a pipeline never runs twice at the same time, and
// records every run. Schedulers, the HTTP API and the CLI all go
// through it, so those rules live in exactly one place.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/run"
)

var (
	// ErrNotFound is returned for an unknown pipeline name.
	ErrNotFound = errors.New("pipeline not found")
	// ErrAlreadyRunning is returned when a run is requested for a
	// pipeline that is still running. Two concurrent "replace" runs on
	// the same table would otherwise race each other's DELETEs.
	ErrAlreadyRunning = errors.New("pipeline is already running")
)

// Executor runs a pipeline once. core.Runner implements it.
type Executor interface {
	Run(ctx context.Context, p core.Pipeline) (core.Stats, error)
}

// Info describes a pipeline for listings: everything but its live
// connectors.
type Info struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Enabled          bool   `json:"enabled"`
	SourceConnection string `json:"source_connection"`
	TargetConnection string `json:"target_connection"`
	TargetTable      string `json:"target_table"`
	Mode             string `json:"mode"`
	ScheduleEvery    string `json:"schedule_every,omitempty"`
}

// Definition pairs a runnable pipeline with its description.
type Definition struct {
	Pipeline core.Pipeline
	Info     Info
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger (default: discard).
func WithLogger(l *slog.Logger) Option { return func(s *Service) { s.log = l } }

// WithIDGenerator replaces UUIDv4 run IDs (for deterministic tests).
func WithIDGenerator(f func() string) Option { return func(s *Service) { s.newID = f } }

// WithClock replaces time.Now (for deterministic tests).
func WithClock(f func() time.Time) Option { return func(s *Service) { s.now = f } }

// WithRunContext sets the parent context of runs started by Trigger.
// Cancelling it aborts (and rolls back) every asynchronous run.
func WithRunContext(ctx context.Context) Option { return func(s *Service) { s.runCtx = ctx } }

// Service runs pipelines by name.
type Service struct {
	defs     map[string]Definition
	order    []string
	exec     Executor
	recorder run.Recorder
	log      *slog.Logger
	newID    func() string
	now      func() time.Time
	runCtx   context.Context

	mu      sync.Mutex
	running map[string]bool
	wg      sync.WaitGroup
}

// New returns a Service for defs.
func New(defs []Definition, exec Executor, recorder run.Recorder, opts ...Option) (*Service, error) {
	s := &Service{
		defs:     make(map[string]Definition, len(defs)),
		exec:     exec,
		recorder: recorder,
		log:      slog.New(slog.DiscardHandler),
		newID:    uuid.NewString,
		now:      time.Now,
		runCtx:   context.Background(),
		running:  map[string]bool{},
	}
	for _, opt := range opts {
		opt(s)
	}
	for _, d := range defs {
		name := d.Pipeline.Name
		if _, dup := s.defs[name]; dup {
			return nil, fmt.Errorf("duplicate pipeline %q", name)
		}
		if err := d.Pipeline.Validate(); err != nil {
			return nil, err
		}
		d.Info.Name = name
		s.defs[name] = d
		s.order = append(s.order, name)
	}
	return s, nil
}

// Pipelines lists every pipeline, in definition order.
func (s *Service) Pipelines() []Info {
	out := make([]Info, len(s.order))
	for i, name := range s.order {
		out[i] = s.defs[name].Info
	}
	return out
}

// Pipeline returns one pipeline's description.
func (s *Service) Pipeline(name string) (Info, bool) {
	d, ok := s.defs[name]
	return d.Info, ok
}

// Execute runs a pipeline and waits for it. The returned Run is final
// (succeeded or failed); the error is the run's failure, if any.
func (s *Service) Execute(ctx context.Context, name string, trigger run.Trigger) (run.Run, error) {
	def, r, err := s.begin(ctx, name, trigger)
	if err != nil {
		return run.Run{}, err
	}
	return s.finish(ctx, def, r)
}

// Trigger starts a pipeline in the background and returns its Run
// right away, in the running state. Use Wait to let background runs
// finish.
func (s *Service) Trigger(name string, trigger run.Trigger) (run.Run, error) {
	def, r, err := s.begin(s.runCtx, name, trigger)
	if err != nil {
		return run.Run{}, err
	}

	s.wg.Go(func() {
		// The failure is already recorded and logged by finish.
		_, _ = s.finish(s.runCtx, def, r)
	})
	return r, nil
}

// Wait blocks until every run started by Trigger has finished, or ctx
// is done.
func (s *Service) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) begin(ctx context.Context, name string, trigger run.Trigger) (Definition, run.Run, error) {
	def, ok := s.defs[name]
	if !ok {
		return Definition{}, run.Run{}, fmt.Errorf("%w: %q", ErrNotFound, name)
	}

	s.mu.Lock()
	if s.running[name] {
		s.mu.Unlock()
		return Definition{}, run.Run{}, fmt.Errorf("%w: %q", ErrAlreadyRunning, name)
	}
	s.running[name] = true
	s.mu.Unlock()

	r := run.Run{
		ID:        s.newID(),
		Pipeline:  name,
		Trigger:   trigger,
		Status:    run.StatusRunning,
		StartedAt: s.now().UTC(),
	}
	s.save(ctx, r)
	s.log.Info("run started", "pipeline", name, "run", r.ID, "trigger", trigger)
	return def, r, nil
}

func (s *Service) finish(ctx context.Context, def Definition, r run.Run) (run.Run, error) {
	defer func() {
		s.mu.Lock()
		delete(s.running, r.Pipeline)
		s.mu.Unlock()
	}()

	stats, err := s.exec.Run(ctx, def.Pipeline)

	finished := s.now().UTC()
	r.FinishedAt = &finished
	r.DurationMS = finished.Sub(r.StartedAt).Milliseconds()
	r.RowsRead, r.RowsWritten, r.Batches = stats.RowsRead, stats.RowsWritten, stats.Batches
	r.Status = run.StatusSucceeded
	if err != nil {
		r.Status = run.StatusFailed
		r.Error = err.Error()
	}
	s.save(context.WithoutCancel(ctx), r)

	attrs := []any{
		"pipeline", r.Pipeline, "run", r.ID, "status", r.Status,
		"rowsRead", r.RowsRead, "rowsWritten", r.RowsWritten, "durationMs", r.DurationMS,
	}
	if err != nil {
		s.log.Error("run failed", append(attrs, "error", err)...)
		return r, err
	}
	s.log.Info("run finished", attrs...)
	return r, nil
}

// save records r. A history store outage must not fail the export
// itself, so errors are logged, not returned.
func (s *Service) save(ctx context.Context, r run.Run) {
	if err := s.recorder.Save(ctx, r); err != nil {
		s.log.Warn("could not record run", "pipeline", r.Pipeline, "run", r.ID, "error", err)
	}
}
