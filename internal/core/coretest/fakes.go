// Package coretest provides in-memory implementations of the core
// ports for tests. They are deliberately simple and fully observable,
// so tests assert on what happened instead of mocking call sequences.
package coretest

import (
	"context"
	"io"
	"sync"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// Source serves a fixed set of rows, BatchSize at a time.
type Source struct {
	Columns []core.Column
	Rows    []core.Row

	// OpenErr is returned by Open; NextErr by Next once FailAfter
	// batches have been served.
	OpenErr   error
	NextErr   error
	FailAfter int

	// Block makes Next wait for ctx cancellation after FailAfter
	// batches instead of returning, to exercise timeouts.
	Block bool

	mu     sync.Mutex
	closed int
}

// Open implements core.Source.
func (s *Source) Open(_ context.Context, q core.Query) (core.Cursor, error) {
	if s.OpenErr != nil {
		return nil, s.OpenErr
	}
	return &cursor{src: s, batchSize: q.BatchSize}, nil
}

// Closed reports how many cursors were closed.
func (s *Source) Closed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

type cursor struct {
	src       *Source
	batchSize int
	pos       int
	served    int
}

func (c *cursor) Columns() []core.Column { return c.src.Columns }

func (c *cursor) Next(ctx context.Context) (core.Batch, error) {
	if (c.src.NextErr != nil || c.src.Block) && c.served >= c.src.FailAfter {
		if c.src.Block {
			<-ctx.Done()
			return core.Batch{}, ctx.Err()
		}
		return core.Batch{}, c.src.NextErr
	}
	if c.pos >= len(c.src.Rows) {
		return core.Batch{}, io.EOF
	}
	end := min(c.pos+c.batchSize, len(c.src.Rows))
	b := core.Batch{Columns: c.src.Columns, Rows: c.src.Rows[c.pos:end]}
	c.pos = end
	c.served++
	return b, nil
}

func (c *cursor) Close() error {
	c.src.mu.Lock()
	defer c.src.mu.Unlock()
	c.src.closed++
	return nil
}

// Sink records everything written through it, and what was committed.
type Sink struct {
	OpenErr   error
	WriteErr  error
	CommitErr error
	// FailOnBatch makes the Nth Write (1-based) return WriteErr.
	FailOnBatch int

	mu         sync.Mutex
	opened     []OpenCall
	pending    []core.Row
	committed  []core.Row
	commits    int
	rollbacks  int
	writeCalls int
}

// OpenCall records the arguments of one Sink.Open call.
type OpenCall struct {
	Target  core.Target
	Columns []core.Column
}

// Open implements core.Sink.
func (s *Sink) Open(_ context.Context, target core.Target, columns []core.Column) (core.Writer, error) {
	if s.OpenErr != nil {
		return nil, s.OpenErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.opened = append(s.opened, OpenCall{Target: target, Columns: columns})
	s.pending = nil
	return &writer{sink: s}, nil
}

// Opened returns every Open call, in order.
func (s *Sink) Opened() []OpenCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]OpenCall(nil), s.opened...)
}

// Committed returns every row of every committed unit of work.
func (s *Sink) Committed() []core.Row {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]core.Row(nil), s.committed...)
}

// Counts returns how many commits and rollbacks happened.
func (s *Sink) Counts() (commits, rollbacks int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commits, s.rollbacks
}

type writer struct {
	sink *Sink
	done bool
}

func (w *writer) Write(_ context.Context, b core.Batch) (int64, error) {
	s := w.sink
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeCalls++
	if s.WriteErr != nil && (s.FailOnBatch == 0 || s.FailOnBatch == s.writeCalls) {
		return 0, s.WriteErr
	}
	s.pending = append(s.pending, b.Rows...)
	return int64(len(b.Rows)), nil
}

func (w *writer) Commit(context.Context) error {
	s := w.sink
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CommitErr != nil {
		return s.CommitErr
	}
	s.committed = append(s.committed, s.pending...)
	s.pending = nil
	s.commits++
	w.done = true
	return nil
}

func (w *writer) Rollback(context.Context) error {
	s := w.sink
	s.mu.Lock()
	defer s.mu.Unlock()
	if w.done {
		return nil
	}
	s.pending = nil
	s.rollbacks++
	w.done = true
	return nil
}

// Rows builds n rows of the shape {i, "row-i"}, handy for batching tests.
func Rows(n int) []core.Row {
	rows := make([]core.Row, n)
	for i := range rows {
		rows[i] = core.Row{i, "row"}
	}
	return rows
}

// Columns is the column set matching Rows.
func Columns() []core.Column {
	return []core.Column{{Name: "id", DatabaseType: "INT4"}, {Name: "name", DatabaseType: "VARCHAR"}}
}
