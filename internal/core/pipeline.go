package core

import (
	"errors"
	"fmt"
	"time"
)

// Pipeline is a fully wired export: where to read, how to reshape,
// where to write. It holds interfaces, not connection settings, so it
// can be built from any configuration format and tested with fakes.
type Pipeline struct {
	Name        string
	Description string

	Source Source
	Query  Query

	// Transformer is optional; nil passes batches through unchanged.
	Transformer Transformer

	Sink   Sink
	Target Target

	// Timeout bounds a single run. Zero means no timeout beyond the
	// caller's context.
	Timeout time.Duration
}

// Validate checks the invariants the Runner relies on.
func (p Pipeline) Validate() error {
	var errs []error
	if p.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if p.Source == nil {
		errs = append(errs, errors.New("source is required"))
	}
	if p.Sink == nil {
		errs = append(errs, errors.New("sink is required"))
	}
	if p.Query.SQL == "" {
		errs = append(errs, errors.New("query is required"))
	}
	if p.Query.BatchSize <= 0 {
		errs = append(errs, fmt.Errorf("batch size must be positive, got %d", p.Query.BatchSize))
	}
	if p.Target.Table == "" {
		errs = append(errs, errors.New("target table is required"))
	}
	if err := p.Target.Mode.Validate(); err != nil {
		errs = append(errs, err)
	}
	if p.Target.Mode == ModeUpsert && len(p.Target.Keys) == 0 {
		errs = append(errs, errors.New("upsert mode requires at least one key column"))
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("pipeline %q: %w", p.Name, err)
	}
	return nil
}
