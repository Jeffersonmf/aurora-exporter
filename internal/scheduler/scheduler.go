// Package scheduler runs jobs periodically. The application depends on
// the Scheduler interface; GoWorkers is the implementation, backed by
// github.com/Jeffersonmf/go-workers, and could be swapped for a cron
// expression engine or a distributed scheduler without touching the
// callers (Dependency Inversion).
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	workermanager "github.com/Jeffersonmf/go-workers/pkg/worker_manager"
)

// Job is one periodic task.
type Job struct {
	Name string
	// Every is the interval between runs. The first run happens as
	// soon as the scheduler starts.
	Every time.Duration
	// Attempts is how many times a failing run is tried per tick.
	Attempts int
	Run      func(ctx context.Context) error
}

// Scheduler starts jobs and lets callers wait for them to stop.
type Scheduler interface {
	// Start schedules every job until ctx is cancelled. It returns
	// immediately.
	Start(ctx context.Context, jobs []Job) error
	// Wait blocks until every job has stopped after ctx cancellation,
	// including any run that was in flight.
	Wait()
}

// ErrSkip is returned (or wrapped) by Job.Run to mean "nothing to do
// this tick": the scheduler neither retries nor reports it as a failure.
var ErrSkip = errors.New("skip")

// GoWorkers schedules each job on its own go-workers Worker, which adds
// retries and panic recovery around every execution.
type GoWorkers struct {
	log *slog.Logger
	wg  sync.WaitGroup
}

// NewGoWorkers returns a GoWorkers scheduler logging to log.
func NewGoWorkers(log *slog.Logger) *GoWorkers {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &GoWorkers{log: log}
}

// Start implements Scheduler.
func (s *GoWorkers) Start(ctx context.Context, jobs []Job) error {
	for _, j := range jobs {
		if j.Every < time.Second {
			return fmt.Errorf("job %q: interval must be at least 1s, got %s", j.Name, j.Every)
		}
	}

	for _, j := range jobs {
		w := workermanager.Worker{
			SourceContext: ctx,
			CronSchedulerConfig: workermanager.CronSchedulerConfig{
				IntervalInSeconds: int(j.Every / time.Second),
				// Sync inside our own goroutine (instead of Async) so
				// Wait can observe when the worker has really stopped.
				TypeOfExecution: workermanager.TypeOfExecutionEnum.Sync,
			},
			MaxRetries: max(j.Attempts, 1),
			Instrumentation: workermanager.Instrumentation{
				FuncName:      j.Name,
				TaskArguments: workermanager.NewTaskParams(),
				FuncDispatcher: func(ctx context.Context, p workermanager.TaskParams) (workermanager.TaskParams, error) {
					err := j.Run(ctx)
					if errors.Is(err, ErrSkip) {
						s.log.Info("scheduled run skipped", "job", j.Name, "reason", err)
						return p, nil
					}
					return p, err
				},
			},
			OnError: func(err *workermanager.TaskError) {
				if ctx.Err() != nil {
					return // shutting down: the cancellation is not a job failure
				}
				s.log.Error("scheduled run failed", "job", j.Name, "attempts", err.Attempt, "error", err.Err)
			},
		}

		s.log.Info("job scheduled", "job", j.Name, "every", j.Every.String())
		s.wg.Go(w.Run)
	}
	return nil
}

// Wait implements Scheduler.
func (s *GoWorkers) Wait() { s.wg.Wait() }
