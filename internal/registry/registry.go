// Package registry maps driver names ("redshift", "postgres", "mysql",
// "jsonl", ...) to the factories that build Sources and Sinks for them.
//
// It is the seam that keeps the rest of the application closed for
// modification: supporting a new database is one RegisterSource or
// RegisterSink call, and nothing that reads configuration, schedules
// runs or serves HTTP needs to know the new driver exists.
package registry

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// Connection is everything a factory needs to build a connector,
// independent of the configuration format it came from.
type Connection struct {
	Name    string
	Driver  string
	DSN     string
	Pool    Pool
	Options map[string]string
}

// Pool tunes database/sql connection pooling. Zero values keep the
// database/sql defaults.
type Pool struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// IntOption parses Options[key] as an int, returning fallback when it
// is not set.
func (c Connection) IntOption(key string, fallback int) (int, error) {
	raw, ok := c.Options[key]
	if !ok || raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("connection %q: option %s: %w", c.Name, key, err)
	}
	return n, nil
}

// SourceFactory builds a Source for a connection.
type SourceFactory func(c Connection) (core.Source, error)

// SinkFactory builds a Sink for a connection.
type SinkFactory func(c Connection) (core.Sink, error)

// Registry holds source and sink factories by driver name.
type Registry struct {
	sources map[string]SourceFactory
	sinks   map[string]SinkFactory
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{sources: map[string]SourceFactory{}, sinks: map[string]SinkFactory{}}
}

// RegisterSource makes driver usable as a pipeline source.
func (r *Registry) RegisterSource(driver string, f SourceFactory) { r.sources[driver] = f }

// RegisterSink makes driver usable as a pipeline target.
func (r *Registry) RegisterSink(driver string, f SinkFactory) { r.sinks[driver] = f }

// NewSource builds the source for c.
func (r *Registry) NewSource(c Connection) (core.Source, error) {
	f, ok := r.sources[c.Driver]
	if !ok {
		return nil, fmt.Errorf("connection %q: driver %q cannot be used as a source (available: %v)", c.Name, c.Driver, r.SourceDrivers())
	}
	return f(c)
}

// NewSink builds the sink for c.
func (r *Registry) NewSink(c Connection) (core.Sink, error) {
	f, ok := r.sinks[c.Driver]
	if !ok {
		return nil, fmt.Errorf("connection %q: driver %q cannot be used as a target (available: %v)", c.Name, c.Driver, r.SinkDrivers())
	}
	return f(c)
}

// HasSource reports whether driver can be read from.
func (r *Registry) HasSource(driver string) bool { _, ok := r.sources[driver]; return ok }

// HasSink reports whether driver can be written to.
func (r *Registry) HasSink(driver string) bool { _, ok := r.sinks[driver]; return ok }

// SourceDrivers lists source driver names, sorted.
func (r *Registry) SourceDrivers() []string { return sortedKeys(r.sources) }

// SinkDrivers lists sink driver names, sorted.
func (r *Registry) SinkDrivers() []string { return sortedKeys(r.sinks) }

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
