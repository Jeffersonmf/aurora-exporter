package app_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/Jeffersonmf/aurora-exporter/internal/app"
	"github.com/Jeffersonmf/aurora-exporter/internal/config"
	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/core/coretest"
	"github.com/Jeffersonmf/aurora-exporter/internal/registry"
	"github.com/Jeffersonmf/aurora-exporter/internal/run"
	"github.com/Jeffersonmf/aurora-exporter/internal/scheduler"
	"github.com/Jeffersonmf/aurora-exporter/internal/transform"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

const doc = `
version: 1
server: {shutdown_timeout: 2s}
connections:
  src: {driver: mem}
  dst: {driver: mem}
pipelines:
  - name: scheduled
    source: {connection: src, query: "select 1"}
    target: {connection: dst, table: t1}
    schedule: {every: 1h}
    transforms: [lowercase_columns]
  - name: disabled
    enabled: false
    source: {connection: src, query: "select 1"}
    target: {connection: dst, table: t2}
    schedule: {every: 1h}
  - name: manual
    source: {connection: src, query: "select 1"}
    target: {connection: dst, table: t3, mode: upsert, keys: [id]}
`

// lifecycle counts Ping/Close calls on a fake connector.
type lifecycle struct {
	pingErr error
	closed  atomic.Int32
}

func (l *lifecycle) Ping(context.Context) error { return l.pingErr }
func (l *lifecycle) Close() error               { l.closed.Add(1); return nil }

type memSource struct {
	*coretest.Source
	*lifecycle
}

type memSink struct {
	*coretest.Sink
	*lifecycle
}

type fixture struct {
	mu      sync.Mutex
	created []*lifecycle
	sinks   []*coretest.Sink
	pingErr error
}

func (f *fixture) registry() *registry.Registry {
	r := registry.New()
	track := func() *lifecycle {
		f.mu.Lock()
		defer f.mu.Unlock()
		l := &lifecycle{pingErr: f.pingErr}
		f.created = append(f.created, l)
		return l
	}
	r.RegisterSource("mem", func(registry.Connection) (core.Source, error) {
		return memSource{Source: &coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(3)}, lifecycle: track()}, nil
	})
	r.RegisterSink("mem", func(registry.Connection) (core.Sink, error) {
		sink := &coretest.Sink{}
		f.mu.Lock()
		f.sinks = append(f.sinks, sink)
		f.mu.Unlock()
		return memSink{Sink: sink, lifecycle: track()}, nil
	})
	return r
}

func build(t *testing.T, f *fixture, yaml string) *app.App {
	t.Helper()
	cfg, err := config.Parse([]byte(yaml), func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatal(err)
	}
	a, err := app.Build(cfg, app.Options{Connectors: f.registry(), Transforms: transform.NewRegistry()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a
}

func TestBuild_WiresPipelinesAndJobs(t *testing.T) {
	t.Parallel()
	f := &fixture{}
	a := build(t, f, doc)

	if got := len(a.Service.Pipelines()); got != 3 {
		t.Fatalf("pipelines = %d, want 3", got)
	}
	if len(a.Jobs) != 1 || a.Jobs[0].Name != "scheduled" || a.Jobs[0].Every != time.Hour {
		t.Fatalf("jobs = %+v, want only the enabled scheduled pipeline", a.Jobs)
	}
	// One source and one sink, shared by every pipeline using them.
	if len(f.created) != 2 {
		t.Fatalf("connectors opened = %d, want 2", len(f.created))
	}

	info, _ := a.Service.Pipeline("scheduled")
	if info.ScheduleEvery != "1h0m0s" || info.TargetTable != "t1" || info.Mode != "replace" {
		t.Fatalf("info = %+v", info)
	}
}

func TestBuild_ExecutesThroughTheWiredConnectors(t *testing.T) {
	t.Parallel()
	f := &fixture{}
	a := build(t, f, doc)

	r, err := a.Service.Execute(context.Background(), "scheduled", run.TriggerCLI)
	if err != nil {
		t.Fatal(err)
	}
	if r.RowsWritten != 3 {
		t.Fatalf("rows written = %d", r.RowsWritten)
	}
	sink := f.sinks[0]
	if got := len(sink.Committed()); got != 3 {
		t.Fatalf("committed = %d", got)
	}
}

func TestBuild_ReportsEveryWiringError(t *testing.T) {
	t.Parallel()
	bad := strings.NewReplacer(
		"dst: {driver: mem}", "dst: {driver: oracle}",
		"transforms: [lowercase_columns]", "transforms: [shout]",
	).Replace(doc)
	cfg, err := config.Parse([]byte(bad), func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatal(err)
	}

	_, err = app.Build(cfg, app.Options{Connectors: (&fixture{}).registry(), Transforms: transform.NewRegistry()})
	if err == nil {
		t.Fatal("expected wiring errors")
	}
	for _, want := range []string{`driver "oracle"`, `unknown transform "shout"`, `pipeline "manual"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in: %v", want, err)
		}
	}
}

func TestReadyAndClose(t *testing.T) {
	t.Parallel()
	f := &fixture{pingErr: errors.New("unreachable")}
	a := build(t, f, doc)

	err := a.Ready(context.Background())
	if err == nil || !strings.Contains(err.Error(), "source src") || !strings.Contains(err.Error(), "target dst") {
		t.Fatalf("Ready = %v, want both connectors reported", err)
	}

	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.created {
		if c.closed.Load() != 1 {
			t.Fatal("connector not closed exactly once")
		}
	}
}

// fakeScheduler runs every job once, synchronously, on Start.
type fakeScheduler struct{ started atomic.Int32 }

func (s *fakeScheduler) Start(ctx context.Context, jobs []scheduler.Job) error {
	for _, j := range jobs {
		s.started.Add(1)
		if err := j.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
func (s *fakeScheduler) Wait() {}

func TestServe_ServesAPIAndShutsDownCleanly(t *testing.T) {
	t.Parallel()
	a := build(t, &fixture{}, doc)

	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	sched := &fakeScheduler{}
	done := make(chan error, 1)
	go func() { done <- a.Serve(ctx, app.ServeOptions{Scheduler: sched, Listener: ln, Version: "test"}) }()

	base := "http://" + ln.Addr().String()
	waitFor(t, base+"/healthz")

	resp, err := client.Post(base+"/v1/pipelines/manual/runs", "application/json", nil) //nolint:noctx // test
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("trigger status = %d", resp.StatusCode)
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Serve = %v", err)
	}
	if sched.started.Load() != 1 {
		t.Fatalf("scheduled jobs started = %d, want 1", sched.started.Load())
	}

	runs, _ := a.Runs.List(context.Background(), run.Filter{})
	if len(runs) != 2 {
		t.Fatalf("runs = %d, want the scheduled and the API run", len(runs))
	}
	for _, r := range runs {
		if r.Status != run.StatusSucceeded {
			t.Errorf("run %s (%s) status = %s; shutdown must let in-flight runs finish", r.ID, r.Pipeline, r.Status)
		}
	}
}

// client never keeps connections around, so these tests control
// exactly which connections exist at shutdown.
var client = &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

func TestServe_IdleTCPConnectionDoesNotBlockShutdown(t *testing.T) {
	t.Parallel()
	a := build(t, &fixture{}, strings.Replace(doc, "shutdown_timeout: 2s", "shutdown_timeout: 200ms", 1))

	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Serve(ctx, app.ServeOptions{Scheduler: &fakeScheduler{}, Listener: ln}) }()
	waitFor(t, "http://"+ln.Addr().String()+"/healthz")

	// What a load balancer's TCP health check looks like: connect, send
	// nothing. net/http waits 5s before treating it as idle.
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve = %v, want a clean stop", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return: an idle connection blocked shutdown")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("shutdown took %s with a 200ms budget", elapsed)
	}
}

func waitFor(t *testing.T, url string) {
	t.Helper()
	for range 100 {
		resp, err := client.Get(url) //nolint:noctx // test
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server at %s never came up", url)
}
