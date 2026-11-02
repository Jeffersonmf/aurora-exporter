package service_test

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/core/coretest"
	"github.com/Jeffersonmf/aurora-exporter/internal/run"
	"github.com/Jeffersonmf/aurora-exporter/internal/service"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// executor is a fake service.Executor: it blocks until release is
// closed (when set) and returns err.
type executor struct {
	err     error
	release chan struct{}
	started chan struct{}
	calls   atomic.Int32
}

func (e *executor) Run(ctx context.Context, _ core.Pipeline) (core.Stats, error) {
	e.calls.Add(1)
	if e.started != nil {
		e.started <- struct{}{}
	}
	if e.release != nil {
		select {
		case <-e.release:
		case <-ctx.Done():
			return core.Stats{}, ctx.Err()
		}
	}
	return core.Stats{RowsRead: 7, RowsWritten: 7, Batches: 1}, e.err
}

func definition(name string) service.Definition {
	return service.Definition{
		Pipeline: core.Pipeline{
			Name:   name,
			Source: &coretest.Source{},
			Query:  core.Query{SQL: "select 1", BatchSize: 10},
			Sink:   &coretest.Sink{},
			Target: core.Target{Table: "t", Mode: core.ModeAppend},
		},
		Info: service.Info{Description: "desc " + name},
	}
}

func newService(t *testing.T, exec service.Executor, store run.Recorder, opts ...service.Option) *service.Service {
	t.Helper()
	var n atomic.Int32
	opts = append([]service.Option{service.WithIDGenerator(func() string { return "run-" + strconv.Itoa(int(n.Add(1))) })}, opts...)
	s, err := service.New([]service.Definition{definition("a"), definition("b")}, exec, store, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestExecute_RecordsSuccess(t *testing.T) {
	t.Parallel()
	store := run.NewMemoryStore(0)
	s := newService(t, &executor{}, store)

	r, err := s.Execute(context.Background(), "a", run.TriggerCLI)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != run.StatusSucceeded || r.RowsWritten != 7 || r.FinishedAt == nil || r.Trigger != run.TriggerCLI {
		t.Fatalf("run = %+v", r)
	}
	stored, _ := store.Get(context.Background(), r.ID)
	if stored.Status != run.StatusSucceeded {
		t.Fatalf("stored status = %s", stored.Status)
	}
}

func TestExecute_RecordsFailure(t *testing.T) {
	t.Parallel()
	boom := errors.New("boom")
	store := run.NewMemoryStore(0)
	s := newService(t, &executor{err: boom}, store)

	r, err := s.Execute(context.Background(), "a", run.TriggerSchedule)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	if r.Status != run.StatusFailed || r.Error != "boom" {
		t.Fatalf("run = %+v", r)
	}
}

func TestExecute_UnknownPipeline(t *testing.T) {
	t.Parallel()
	s := newService(t, &executor{}, run.NewMemoryStore(0))
	if _, err := s.Execute(context.Background(), "zzz", run.TriggerAPI); !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestTrigger_RejectsConcurrentRunOfSamePipeline(t *testing.T) {
	t.Parallel()
	exec := &executor{release: make(chan struct{}), started: make(chan struct{}, 2)}
	store := run.NewMemoryStore(0)
	s := newService(t, exec, store)

	first, err := s.Trigger("a", run.TriggerAPI)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != run.StatusRunning {
		t.Fatalf("Trigger must return the run while it is running, got %s", first.Status)
	}
	<-exec.started

	if _, err := s.Trigger("a", run.TriggerAPI); !errors.Is(err, service.ErrAlreadyRunning) {
		t.Fatalf("second trigger: err = %v, want ErrAlreadyRunning", err)
	}
	if _, err := s.Execute(context.Background(), "a", run.TriggerSchedule); !errors.Is(err, service.ErrAlreadyRunning) {
		t.Fatalf("scheduled run during API run: err = %v, want ErrAlreadyRunning", err)
	}
	// A different pipeline is not blocked.
	if _, err := s.Trigger("b", run.TriggerAPI); err != nil {
		t.Fatalf("other pipeline: %v", err)
	}
	<-exec.started

	close(exec.release)
	if err := s.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	got, _ := store.Get(context.Background(), first.ID)
	if got.Status != run.StatusSucceeded {
		t.Fatalf("status after Wait = %s", got.Status)
	}
	// The slot is released: the pipeline can run again.
	if _, err := s.Execute(context.Background(), "a", run.TriggerCLI); err != nil {
		t.Fatalf("run after completion: %v", err)
	}
}

func TestWait_HonorsDeadlineAndRunContextCancels(t *testing.T) {
	t.Parallel()
	runCtx, cancelRuns := context.WithCancel(context.Background())
	exec := &executor{release: make(chan struct{}), started: make(chan struct{}, 1)}
	store := run.NewMemoryStore(0)
	s := newService(t, exec, store, service.WithRunContext(runCtx))

	r, _ := s.Trigger("a", run.TriggerAPI)
	<-exec.started

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := s.Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Wait = %v, want deadline exceeded while a run is in flight", err)
	}

	cancelRuns()
	if err := s.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Get(context.Background(), r.ID)
	if got.Status != run.StatusFailed {
		t.Fatalf("cancelled run status = %s, want failed", got.Status)
	}
}

func TestNew_RejectsInvalidDefinitions(t *testing.T) {
	t.Parallel()
	if _, err := service.New([]service.Definition{definition("a"), definition("a")}, &executor{}, run.NewMemoryStore(0)); err == nil {
		t.Fatal("expected duplicate pipeline error")
	}
	bad := definition("bad")
	bad.Pipeline.Query.SQL = ""
	if _, err := service.New([]service.Definition{bad}, &executor{}, run.NewMemoryStore(0)); err == nil {
		t.Fatal("expected invalid pipeline error")
	}
}

func TestPipelines_KeepsDefinitionOrder(t *testing.T) {
	t.Parallel()
	s := newService(t, &executor{}, run.NewMemoryStore(0))
	got := s.Pipelines()
	if len(got) != 2 || got[0].Name != "a" || got[1].Name != "b" || got[0].Description != "desc a" {
		t.Fatalf("pipelines = %+v", got)
	}
	if _, ok := s.Pipeline("b"); !ok {
		t.Fatal("Pipeline(b) not found")
	}
}

type failingRecorder struct{}

func (failingRecorder) Save(context.Context, run.Run) error { return errors.New("store down") }

func TestExecute_HistoryOutageDoesNotFailTheRun(t *testing.T) {
	t.Parallel()
	s := newService(t, &executor{}, failingRecorder{})
	if _, err := s.Execute(context.Background(), "a", run.TriggerCLI); err != nil {
		t.Fatalf("err = %v, a recorder failure must only be logged", err)
	}
}
