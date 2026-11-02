package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/Jeffersonmf/aurora-exporter/internal/scheduler"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestGoWorkers_RunsImmediatelyAndStopsOnCancel(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	ran := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())

	s := scheduler.NewGoWorkers(nil)
	err := s.Start(ctx, []scheduler.Job{{
		Name: "job", Every: time.Hour, Attempts: 1,
		Run: func(context.Context) error {
			calls.Add(1)
			ran <- struct{}{}
			return nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-ran:
	case <-time.After(2 * time.Second):
		t.Fatal("job did not run on start")
	}
	cancel()
	s.Wait() // goleak asserts the worker goroutine is gone
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestGoWorkers_RetriesUpToAttempts(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := scheduler.NewGoWorkers(nil)
	_ = s.Start(ctx, []scheduler.Job{{
		Name: "flaky", Every: time.Hour, Attempts: 3,
		Run: func(context.Context) error {
			if calls.Add(1) == 3 {
				close(done)
				return nil
			}
			return errors.New("transient")
		},
	}})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("calls = %d, want the 3rd attempt to happen", calls.Load())
	}
	cancel()
	s.Wait()
}

func TestGoWorkers_SkipIsNotRetried(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	ran := make(chan struct{}, 10)
	ctx, cancel := context.WithCancel(context.Background())

	s := scheduler.NewGoWorkers(nil)
	_ = s.Start(ctx, []scheduler.Job{{
		Name: "busy", Every: time.Hour, Attempts: 5,
		Run: func(context.Context) error {
			calls.Add(1)
			ran <- struct{}{}
			return fmt.Errorf("still running: %w", scheduler.ErrSkip)
		},
	}})

	<-ran
	time.Sleep(50 * time.Millisecond) // room for (unwanted) retries
	cancel()
	s.Wait()
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, a skipped run must not be retried", calls.Load())
	}
}

func TestGoWorkers_RejectsSubSecondInterval(t *testing.T) {
	t.Parallel()
	s := scheduler.NewGoWorkers(nil)
	err := s.Start(context.Background(), []scheduler.Job{{Name: "x", Every: time.Millisecond, Run: func(context.Context) error { return nil }}})
	if err == nil {
		t.Fatal("expected an interval validation error")
	}
}
