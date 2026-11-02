package core

import "context"

// Query is what a pipeline asks its Source for.
type Query struct {
	// SQL is the statement to run on the source. Connectors that are not
	// SQL-based may interpret it however they document.
	SQL string
	// BatchSize caps how many rows a single Cursor.Next call returns.
	BatchSize int
}

// Source opens result sets. It is the extraction port: Redshift is one
// implementation, any database/sql driver can be another.
type Source interface {
	Open(ctx context.Context, q Query) (Cursor, error)
}

// Cursor iterates over one result set in batches.
type Cursor interface {
	// Columns describes the result set. It is available right after
	// Open, before any row is read, so a Sink can be prepared even
	// when the result is empty (an empty "replace" must still clear
	// the target).
	Columns() []Column
	// Next returns the next batch of at most Query.BatchSize rows, or
	// io.EOF once the result set is exhausted. The caller owns the
	// returned batch: transformers may modify its rows in place.
	Next(ctx context.Context) (Batch, error)
	Close() error
}

// Sink opens writers. It is the load port.
type Sink interface {
	// Open starts a unit of work against target for rows shaped like
	// columns. Nothing written through the returned Writer is visible
	// to readers of the target until Commit succeeds, when the Sink
	// supports it (SQL sinks do; see each adapter's documentation).
	Open(ctx context.Context, target Target, columns []Column) (Writer, error)
}

// Writer is one unit of work opened by a Sink.
type Writer interface {
	// Write stores every row of b, returning how many were written.
	Write(ctx context.Context, b Batch) (int64, error)
	Commit(ctx context.Context) error
	// Rollback discards the unit of work. Calling it after a
	// successful Commit is a no-op, so callers can always defer it.
	Rollback(ctx context.Context) error
}

// Transformer reshapes batches between Source and Sink. It is called
// once with a schema-only batch (no rows) before any data, then once
// per data batch, and must produce the same columns for both.
type Transformer interface {
	Transform(b Batch) (Batch, error)
}

// TransformerFunc adapts a plain function to the Transformer interface.
type TransformerFunc func(Batch) (Batch, error)

// Transform calls f(b).
func (f TransformerFunc) Transform(b Batch) (Batch, error) { return f(b) }

// Pinger is an optional capability: connectors that can check their
// connectivity implement it, and readiness probes use it. It is a
// separate interface (Interface Segregation) because a Source or Sink
// that cannot be pinged, like a local file, should not be forced to
// fake an implementation.
type Pinger interface {
	Ping(ctx context.Context) error
}
