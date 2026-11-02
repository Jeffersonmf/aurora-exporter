// Package app is the composition root: it turns a validated
// configuration into connectors, transformers, pipelines, a service
// and scheduler jobs. It is the only package that knows every concrete
// type; everything it wires together only knows interfaces.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/Jeffersonmf/aurora-exporter/internal/config"
	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/registry"
	"github.com/Jeffersonmf/aurora-exporter/internal/run"
	"github.com/Jeffersonmf/aurora-exporter/internal/scheduler"
	"github.com/Jeffersonmf/aurora-exporter/internal/service"
	"github.com/Jeffersonmf/aurora-exporter/internal/transform"
)

// Options are the extension points of Build: pass a registry with
// extra drivers or transformers to support them without editing this
// package.
type Options struct {
	Connectors *registry.Registry
	Transforms *transform.Registry
	Logger     *slog.Logger
	// Runs stores run history; nil uses an in-memory store.
	Runs run.Store
}

// App is a fully wired aurora-exporter.
type App struct {
	Config  *config.Config
	Service *service.Service
	Runs    run.Store
	// Jobs are the scheduled pipelines, ready for a Scheduler.
	Jobs []scheduler.Job

	log        *slog.Logger
	connectors []namedConnector
	runCtx     context.Context
	cancelRuns context.CancelFunc
}

type namedConnector struct {
	name string
	impl any
}

// Build wires cfg. It opens connection pools but does not connect:
// database/sql connects lazily, so Build succeeds offline and Ready
// reports reachability separately.
func Build(cfg *config.Config, o Options) (_ *App, err error) {
	if o.Connectors == nil || o.Transforms == nil {
		return nil, errors.New("app: connector and transform registries are required")
	}
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}
	if o.Runs == nil {
		o.Runs = run.NewMemoryStore(0)
	}

	a := &App{Config: cfg, Runs: o.Runs, log: o.Logger}
	a.runCtx, a.cancelRuns = context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			_ = a.Close()
		}
	}()

	b := builder{cfg: cfg, opts: o, sources: map[string]core.Source{}, sinks: map[string]core.Sink{}}
	defs, err := b.definitions()
	a.connectors = b.opened
	if err != nil {
		return nil, err
	}

	a.Service, err = service.New(defs, core.Runner{Logger: o.Logger}, o.Runs,
		service.WithLogger(o.Logger),
		service.WithRunContext(a.runCtx),
	)
	if err != nil {
		return nil, err
	}
	a.Jobs = a.jobs()
	return a, nil
}

func (a *App) jobs() []scheduler.Job {
	var jobs []scheduler.Job
	for _, p := range a.Config.Pipelines {
		if !p.IsEnabled() || p.Schedule.Every == 0 {
			continue
		}
		name := p.Name
		jobs = append(jobs, scheduler.Job{
			Name:     name,
			Every:    p.Schedule.Every.Std(),
			Attempts: *p.Attempts,
			// The scheduler's context only governs ticking. The run
			// itself uses runCtx, so a shutdown stops scheduling at
			// once but lets an in-flight run finish within the
			// shutdown timeout instead of rolling it back immediately.
			Run: func(context.Context) error {
				_, err := a.Service.Execute(a.runCtx, name, run.TriggerSchedule)
				if errors.Is(err, service.ErrAlreadyRunning) {
					return fmt.Errorf("%w: %w", scheduler.ErrSkip, err)
				}
				return err
			},
		})
	}
	return jobs
}

// Ready pings every connector that supports it and reports all
// unreachable ones.
func (a *App) Ready(ctx context.Context) error {
	var errs []error
	for _, c := range a.connectors {
		if p, ok := c.impl.(core.Pinger); ok {
			if err := p.Ping(ctx); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", c.name, err))
			}
		}
	}
	return errors.Join(errs...)
}

// Close cancels background runs and releases every connector.
func (a *App) Close() error {
	a.cancelRuns()
	var errs []error
	for _, c := range a.connectors {
		if cl, ok := c.impl.(io.Closer); ok {
			if err := cl.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close %s: %w", c.name, err))
			}
		}
	}
	return errors.Join(errs...)
}

// builder turns configuration into service definitions, opening each
// connection at most once per role.
type builder struct {
	cfg     *config.Config
	opts    Options
	sources map[string]core.Source
	sinks   map[string]core.Sink
	opened  []namedConnector
}

func (b *builder) definitions() ([]service.Definition, error) {
	var (
		defs []service.Definition
		errs []error
	)
	for _, p := range b.cfg.Pipelines {
		def, err := b.definition(p)
		if err != nil {
			errs = append(errs, fmt.Errorf("pipeline %q: %w", p.Name, err))
			continue
		}
		defs = append(defs, def)
	}
	return defs, errors.Join(errs...)
}

func (b *builder) definition(p config.Pipeline) (service.Definition, error) {
	src, srcErr := b.source(p.Source.Connection)
	sink, sinkErr := b.sink(p.Target.Connection)
	tr, trErr := b.transformer(p.Transforms)
	if err := errors.Join(srcErr, sinkErr, trErr); err != nil {
		return service.Definition{}, err
	}

	pipeline := core.Pipeline{
		Name:        p.Name,
		Description: p.Description,
		Source:      src,
		Query:       core.Query{SQL: p.Source.Query, BatchSize: p.BatchSize},
		Transformer: tr,
		Sink:        sink,
		Target:      core.Target{Table: p.Target.Table, Mode: core.WriteMode(p.Target.Mode), Keys: p.Target.Keys},
		Timeout:     p.Timeout.Std(),
	}
	info := service.Info{
		Description:      p.Description,
		Enabled:          p.IsEnabled(),
		SourceConnection: p.Source.Connection,
		TargetConnection: p.Target.Connection,
		TargetTable:      p.Target.Table,
		Mode:             p.Target.Mode,
	}
	if p.Schedule.Every > 0 {
		info.ScheduleEvery = p.Schedule.Every.Std().String()
	}
	return service.Definition{Pipeline: pipeline, Info: info}, nil
}

func (b *builder) source(name string) (core.Source, error) {
	if s, ok := b.sources[name]; ok {
		return s, nil
	}
	s, err := b.opts.Connectors.NewSource(b.connection(name))
	if err != nil {
		return nil, err
	}
	b.sources[name] = s
	b.opened = append(b.opened, namedConnector{name: "source " + name, impl: s})
	return s, nil
}

func (b *builder) sink(name string) (core.Sink, error) {
	if s, ok := b.sinks[name]; ok {
		return s, nil
	}
	s, err := b.opts.Connectors.NewSink(b.connection(name))
	if err != nil {
		return nil, err
	}
	b.sinks[name] = s
	b.opened = append(b.opened, namedConnector{name: "target " + name, impl: s})
	return s, nil
}

func (b *builder) connection(name string) registry.Connection {
	c := b.cfg.Connections[name]
	return registry.Connection{
		Name:   name,
		Driver: c.Driver,
		DSN:    c.DSN,
		Pool: registry.Pool{
			MaxOpenConns:    c.MaxOpenConns,
			MaxIdleConns:    c.MaxIdleConns,
			ConnMaxLifetime: c.ConnMaxLifetime.Std(),
			ConnMaxIdleTime: c.ConnMaxIdleTime.Std(),
		},
		Options: c.Options,
	}
}

func (b *builder) transformer(specs []config.Transform) (core.Transformer, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	chain := make([]core.Transformer, 0, len(specs))
	var errs []error
	for _, s := range specs {
		var params transform.Decoder
		if s.Params != nil {
			params = s.Params
		}
		t, err := b.opts.Transforms.Build(s.Type, params)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		chain = append(chain, t)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return transform.Chain(chain...), nil
}
