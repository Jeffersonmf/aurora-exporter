package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/Jeffersonmf/aurora-exporter/internal/httpapi"
	"github.com/Jeffersonmf/aurora-exporter/internal/scheduler"
)

// ServeOptions configures Serve.
type ServeOptions struct {
	Scheduler scheduler.Scheduler
	Version   string
	// Listener overrides server.addr (tests listen on a random port).
	Listener net.Listener
}

// Serve schedules every job and serves the HTTP API until ctx is
// cancelled, then shuts down in order:
//
//  1. stop scheduling new runs and stop accepting HTTP requests;
//  2. wait up to server.shutdown_timeout for in-flight runs;
//  3. cancel whatever is still running (its transaction rolls back).
func (a *App) Serve(ctx context.Context, o ServeOptions) error {
	ln := o.Listener
	if ln == nil {
		var err error
		if ln, err = (&net.ListenConfig{}).Listen(ctx, "tcp", a.Config.Server.Addr); err != nil {
			return err
		}
	}

	if err := o.Scheduler.Start(ctx, a.Jobs); err != nil {
		_ = ln.Close()
		return err
	}

	srv := &http.Server{
		Handler: httpapi.NewHandler(httpapi.Deps{
			Catalog:   a.Service,
			Triggerer: a.Service,
			Runs:      a.Runs,
			Readiness: a,
			Logger:    a.log,
			Version:   o.Version,
			AuthToken: a.Config.Server.AuthToken,
		}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()
	a.log.Info("aurora-exporter serving",
		"addr", ln.Addr().String(),
		"pipelines", len(a.Config.Pipelines),
		"scheduled", len(a.Jobs),
	)

	var err error
	select {
	case <-ctx.Done():
	case err = <-serveErr:
	}

	a.log.Info("shutting down", "timeout", a.Config.Server.ShutdownTimeout.Std().String())
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.Config.Server.ShutdownTimeout.Std())
	defer cancel()

	// HTTP and runs drain concurrently, sharing one deadline.
	httpDone := make(chan struct{})
	go func(ctx context.Context) {
		defer close(httpDone)
		if serr := srv.Shutdown(ctx); serr != nil {
			// net/http only treats a connection that never sent a
			// request (a load balancer's TCP health check, a client's
			// pre-dial) as idle after 5s, so Shutdown can time out with
			// no request in flight at all. That must not hold the
			// process or turn a clean stop into a failure.
			a.log.Warn("closing remaining HTTP connections", "reason", serr)
			_ = srv.Close()
		}
	}(shutdownCtx)

	drained := make(chan struct{})
	go func(ctx context.Context) {
		o.Scheduler.Wait()
		_ = a.Service.Wait(ctx)
		close(drained)
	}(context.WithoutCancel(shutdownCtx))
	select {
	case <-drained:
	case <-shutdownCtx.Done():
		a.log.Warn("shutdown timeout reached, cancelling in-flight runs")
		a.cancelRuns()
		<-drained
	}
	<-httpDone

	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	a.log.Info("aurora-exporter stopped")
	return err
}
