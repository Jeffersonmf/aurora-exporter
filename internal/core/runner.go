package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"
)

// defaultPrefetch is how many batches the reader may get ahead of the
// writer when Runner.Prefetch is zero.
const defaultPrefetch = 2

// Stats summarizes one pipeline run.
type Stats struct {
	RowsRead    int64
	RowsWritten int64
	Batches     int
	Duration    time.Duration
}

// Runner executes pipelines. Reading and writing overlap: a reader
// goroutine fetches the next batch from the Source while the current
// one is being written, bounded by Prefetch so memory stays at
// (Prefetch+1) x BatchSize rows no matter how large the table is.
//
// A run is all-or-nothing from the target's point of view: the Writer
// is committed only after the Source is exhausted without errors, and
// rolled back on any failure, including cancellation.
type Runner struct {
	Logger   *slog.Logger
	Prefetch int
}

// Run executes p once.
func (r Runner) Run(ctx context.Context, p Pipeline) (stats Stats, err error) {
	if err := p.Validate(); err != nil {
		return stats, err
	}
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}

	start := time.Now()
	defer func() { stats.Duration = time.Since(start) }()

	cur, err := p.Source.Open(ctx, p.Query)
	if err != nil {
		return stats, fmt.Errorf("open source: %w", err)
	}
	defer func() {
		if cerr := cur.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("close source: %w", cerr))
		}
	}()

	transform := transformerOrIdentity(p.Transformer)
	schema, err := transform.Transform(Batch{Columns: cur.Columns()})
	if err != nil {
		return stats, fmt.Errorf("transform schema: %w", err)
	}

	w, err := p.Sink.Open(ctx, p.Target, schema.Columns)
	if err != nil {
		return stats, fmt.Errorf("open sink: %w", err)
	}
	defer func() {
		if err == nil {
			return
		}
		// The run's ctx may be the very reason we are rolling back
		// (timeout, shutdown); the rollback itself must still reach
		// the database.
		if rerr := w.Rollback(context.WithoutCancel(ctx)); rerr != nil {
			err = errors.Join(err, fmt.Errorf("rollback: %w", rerr))
		}
	}()

	if err = r.pump(ctx, p.Name, cur, transform, w, &stats); err != nil {
		return stats, err
	}
	if err = w.Commit(ctx); err != nil {
		return stats, fmt.Errorf("commit: %w", err)
	}
	return stats, nil
}

// pump moves every batch from cur to w. It returns only after the
// reader goroutine has exited, so the caller can close the cursor
// without racing a concurrent Next.
func (r Runner) pump(ctx context.Context, name string, cur Cursor, t Transformer, w Writer, stats *Stats) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	batches := make(chan Batch, r.prefetch())
	readErr := make(chan error, 1)

	go func() {
		defer close(batches)
		for {
			b, err := cur.Next(ctx)
			if errors.Is(err, io.EOF) {
				readErr <- nil
				return
			}
			if err != nil {
				readErr <- fmt.Errorf("read source: %w", err)
				return
			}
			select {
			case batches <- b:
			case <-ctx.Done():
				readErr <- ctx.Err()
				return
			}
		}
	}()

	writeErr := r.drain(ctx, name, batches, t, w, stats)
	if writeErr != nil {
		cancel()
		for range batches { //nolint:revive // drain so the reader can exit
		}
		// The reader almost always reports our own cancellation here;
		// anything else is a second, independent failure worth keeping.
		if rerr := <-readErr; rerr != nil && !errors.Is(rerr, context.Canceled) {
			return errors.Join(writeErr, rerr)
		}
		return writeErr
	}
	return <-readErr
}

func (r Runner) drain(ctx context.Context, name string, batches <-chan Batch, t Transformer, w Writer, stats *Stats) error {
	log := r.logger()
	for b := range batches {
		stats.Batches++
		stats.RowsRead += int64(b.Len())
		if b.Len() == 0 {
			continue
		}

		out, err := t.Transform(b)
		if err != nil {
			return fmt.Errorf("transform batch %d: %w", stats.Batches, err)
		}

		n, err := w.Write(ctx, out)
		stats.RowsWritten += n
		if err != nil {
			return fmt.Errorf("write batch %d: %w", stats.Batches, err)
		}

		log.Debug("batch written",
			"pipeline", name,
			"batch", stats.Batches,
			"rows", n,
			"rowsWritten", stats.RowsWritten,
		)
	}
	return nil
}

func (r Runner) prefetch() int {
	if r.Prefetch > 0 {
		return r.Prefetch
	}
	return defaultPrefetch
}

func (r Runner) logger() *slog.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return slog.New(slog.DiscardHandler)
}

func transformerOrIdentity(t Transformer) Transformer {
	if t != nil {
		return t
	}
	return TransformerFunc(func(b Batch) (Batch, error) { return b, nil })
}
