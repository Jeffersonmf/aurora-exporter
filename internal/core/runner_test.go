package core_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/core/coretest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func pipeline(src core.Source, sink core.Sink) core.Pipeline {
	return core.Pipeline{
		Name:   "test",
		Source: src,
		Query:  core.Query{SQL: "select 1", BatchSize: 3},
		Sink:   sink,
		Target: core.Target{Table: "t", Mode: core.ModeReplace},
	}
}

func TestRunner_StreamsEveryBatchAndCommitsOnce(t *testing.T) {
	t.Parallel()
	src := &coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(10)}
	sink := &coretest.Sink{}

	stats, err := core.Runner{}.Run(context.Background(), pipeline(src, sink))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if stats.RowsRead != 10 || stats.RowsWritten != 10 || stats.Batches != 4 {
		t.Fatalf("stats = %+v, want 10 read, 10 written, 4 batches", stats)
	}
	if got := len(sink.Committed()); got != 10 {
		t.Fatalf("committed %d rows, want 10", got)
	}
	if commits, rollbacks := sink.Counts(); commits != 1 || rollbacks != 0 {
		t.Fatalf("commits=%d rollbacks=%d, want 1/0", commits, rollbacks)
	}
	if src.Closed() != 1 {
		t.Fatalf("cursor closed %d times, want 1", src.Closed())
	}
}

// An empty result must still open and commit the writer: in replace
// mode that is what clears the target table.
func TestRunner_EmptyResultStillCommits(t *testing.T) {
	t.Parallel()
	src := &coretest.Source{Columns: coretest.Columns()}
	sink := &coretest.Sink{}

	if _, err := (core.Runner{}).Run(context.Background(), pipeline(src, sink)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(sink.Opened()) != 1 {
		t.Fatalf("sink opened %d times, want 1", len(sink.Opened()))
	}
	if commits, _ := sink.Counts(); commits != 1 {
		t.Fatalf("commits = %d, want 1", commits)
	}
}

func TestRunner_WriteErrorRollsBackAndStopsReader(t *testing.T) {
	t.Parallel()
	boom := errors.New("disk full")
	src := &coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(100)}
	sink := &coretest.Sink{WriteErr: boom, FailOnBatch: 2}

	_, err := core.Runner{Prefetch: 1}.Run(context.Background(), pipeline(src, sink))
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if commits, rollbacks := sink.Counts(); commits != 0 || rollbacks != 1 {
		t.Fatalf("commits=%d rollbacks=%d, want 0/1", commits, rollbacks)
	}
	if len(sink.Committed()) != 0 {
		t.Fatal("rows were committed despite the failure")
	}
	// goleak (TestMain) asserts the reader goroutine exited.
}

func TestRunner_ReadErrorRollsBack(t *testing.T) {
	t.Parallel()
	boom := errors.New("connection reset")
	src := &coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(10), NextErr: boom, FailAfter: 2}
	sink := &coretest.Sink{}

	_, err := core.Runner{}.Run(context.Background(), pipeline(src, sink))
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
	if commits, rollbacks := sink.Counts(); commits != 0 || rollbacks != 1 {
		t.Fatalf("commits=%d rollbacks=%d, want 0/1", commits, rollbacks)
	}
}

func TestRunner_TimeoutRollsBack(t *testing.T) {
	t.Parallel()
	src := &coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(10), Block: true, FailAfter: 1}
	sink := &coretest.Sink{}
	p := pipeline(src, sink)
	p.Timeout = 20 * time.Millisecond

	_, err := core.Runner{}.Run(context.Background(), p)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want deadline exceeded", err)
	}
	if _, rollbacks := sink.Counts(); rollbacks != 1 {
		t.Fatalf("rollbacks = %d, want 1", rollbacks)
	}
}

func TestRunner_CommitErrorIsReported(t *testing.T) {
	t.Parallel()
	boom := errors.New("serialization failure")
	src := &coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(2)}
	sink := &coretest.Sink{CommitErr: boom}

	_, err := core.Runner{}.Run(context.Background(), pipeline(src, sink))
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestRunner_OpenErrorsAreWrapped(t *testing.T) {
	t.Parallel()
	boom := errors.New("refused")

	_, err := core.Runner{}.Run(context.Background(), pipeline(&coretest.Source{OpenErr: boom}, &coretest.Sink{}))
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "open source") {
		t.Fatalf("source open: err = %v", err)
	}

	src := &coretest.Source{Columns: coretest.Columns()}
	_, err = core.Runner{}.Run(context.Background(), pipeline(src, &coretest.Sink{OpenErr: boom}))
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "open sink") {
		t.Fatalf("sink open: err = %v", err)
	}
	if src.Closed() != 1 {
		t.Fatal("cursor must be closed even when the sink fails to open")
	}
}

func TestRunner_TransformerShapesSchemaAndRows(t *testing.T) {
	t.Parallel()
	upper := core.TransformerFunc(func(b core.Batch) (core.Batch, error) {
		cols := make([]core.Column, len(b.Columns))
		for i, c := range b.Columns {
			cols[i] = core.Column{Name: strings.ToUpper(c.Name)}
		}
		return core.Batch{Columns: cols, Rows: b.Rows}, nil
	})
	sink := &coretest.Sink{}
	p := pipeline(&coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(1)}, sink)
	p.Transformer = upper

	if _, err := (core.Runner{}).Run(context.Background(), p); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := sink.Opened()[0].Columns[0].Name; got != "ID" {
		t.Fatalf("sink opened with column %q, want the transformed name ID", got)
	}
}

func TestRunner_RejectsInvalidPipeline(t *testing.T) {
	t.Parallel()
	p := pipeline(&coretest.Source{}, &coretest.Sink{})
	p.Target = core.Target{Table: "t", Mode: core.ModeUpsert}

	_, err := core.Runner{}.Run(context.Background(), p)
	if err == nil || !strings.Contains(err.Error(), "requires at least one key") {
		t.Fatalf("err = %v, want upsert-without-keys validation error", err)
	}
}
