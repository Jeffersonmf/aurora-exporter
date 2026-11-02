package cli

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Jeffersonmf/aurora-exporter/internal/app"
	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/run"
	"github.com/Jeffersonmf/aurora-exporter/internal/scheduler"
	"github.com/Jeffersonmf/aurora-exporter/internal/transform"
	"github.com/Jeffersonmf/aurora-exporter/internal/version"
)

// --- serve -----------------------------------------------------------

type serveCmd struct{}

func (serveCmd) Name() string { return "serve" }
func (serveCmd) Synopsis() string {
	return "run scheduled pipelines and the HTTP API until SIGINT/SIGTERM"
}

func (c serveCmd) Run(ctx context.Context, env Env, args []string) int {
	fs := newFlagSet(env, c.Name(), c.Synopsis())
	path := configFlag(fs, env)
	if code, stop := parseFlags(fs, args); stop {
		return code
	}

	a, log, err := bootstrap(env, *path)
	if err != nil {
		fmt.Fprintln(env.Stderr, err)
		return ExitUsage
	}
	defer func() { _ = a.Close() }()

	err = a.Serve(ctx, app.ServeOptions{Scheduler: scheduler.NewGoWorkers(log), Version: version.Version})
	if err != nil {
		log.Error("server stopped", "error", err)
		return ExitFailure
	}
	return ExitOK
}

// --- run -------------------------------------------------------------

type runCmd struct{}

func (runCmd) Name() string     { return "run" }
func (runCmd) Synopsis() string { return "run pipelines once and exit (non-zero if any failed)" }

func (c runCmd) Run(ctx context.Context, env Env, args []string) int {
	fs := newFlagSet(env, c.Name()+" [flags] [pipeline ...]", c.Synopsis())
	path := configFlag(fs, env)
	all := fs.Bool("all", false, "run every enabled pipeline")
	if code, stop := parseFlags(fs, args); stop {
		return code
	}

	names := fs.Args()
	if len(names) == 0 && !*all {
		fmt.Fprintln(env.Stderr, "name at least one pipeline, or pass -all")
		fs.Usage()
		return ExitUsage
	}

	a, _, err := bootstrap(env, *path)
	if err != nil {
		fmt.Fprintln(env.Stderr, err)
		return ExitUsage
	}
	defer func() { _ = a.Close() }()

	if *all {
		for _, p := range a.Service.Pipelines() {
			if p.Enabled && !slices.Contains(names, p.Name) {
				names = append(names, p.Name)
			}
		}
	}
	for _, n := range names {
		if _, ok := a.Service.Pipeline(n); !ok {
			fmt.Fprintf(env.Stderr, "unknown pipeline %q\n", n)
			return ExitUsage
		}
	}

	// Pipelines run one after another: they usually share the same
	// source cluster, and serial runs keep its load predictable.
	results := make([]run.Run, 0, len(names))
	failed := false
	for _, n := range names {
		r, err := a.Service.Execute(ctx, n, run.TriggerCLI)
		if err != nil {
			failed = true
			if r.ID == "" { // never started (e.g. already running)
				r = run.Run{Pipeline: n, Status: run.StatusFailed, Error: err.Error()}
			}
		}
		results = append(results, r)
		if ctx.Err() != nil {
			break
		}
	}

	printRuns(env, results)
	if failed {
		return ExitFailure
	}
	return ExitOK
}

func printRuns(env Env, runs []run.Run) {
	tw := tabwriter.NewWriter(env.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PIPELINE\tSTATUS\tROWS\tDURATION\tERROR")
	for _, r := range runs {
		d := (time.Duration(r.DurationMS) * time.Millisecond).String()
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\t%s\n", r.Pipeline, r.Status, r.RowsWritten, d, r.Error)
	}
	_ = tw.Flush()
}

// --- validate --------------------------------------------------------

type validateCmd struct{}

func (validateCmd) Name() string { return "validate" }
func (validateCmd) Synopsis() string {
	return "check the configuration (and, with -ping, connectivity) without running anything"
}

func (c validateCmd) Run(ctx context.Context, env Env, args []string) int {
	fs := newFlagSet(env, c.Name(), c.Synopsis())
	path := configFlag(fs, env)
	ping := fs.Bool("ping", false, "also connect to every database")
	if code, stop := parseFlags(fs, args); stop {
		return code
	}

	a, _, err := bootstrap(env, *path)
	if err != nil {
		fmt.Fprintf(env.Stderr, "configuration invalid:\n%v\n", err)
		return ExitUsage
	}
	defer func() { _ = a.Close() }()

	if *ping {
		pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if err := a.Ready(pingCtx); err != nil {
			fmt.Fprintf(env.Stderr, "connectivity check failed:\n%v\n", err)
			return ExitFailure
		}
	}

	fmt.Fprintf(env.Stdout, "%s: OK (%d pipelines, %d scheduled, %d connections)\n",
		*path, len(a.Config.Pipelines), len(a.Jobs), len(a.Config.Connections))
	return ExitOK
}

// --- drivers ---------------------------------------------------------

type driversCmd struct{}

func (driversCmd) Name() string { return "drivers" }
func (driversCmd) Synopsis() string {
	return "list supported sources, targets, write modes and transforms"
}

func (driversCmd) Run(_ context.Context, env Env, _ []string) int {
	reg := env.Connectors(env.Stdout)
	modes := make([]string, 0, len(core.WriteModes()))
	for _, m := range core.WriteModes() {
		modes = append(modes, string(m))
	}
	fmt.Fprintf(env.Stdout, "sources:     %s\n", strings.Join(reg.SourceDrivers(), ", "))
	fmt.Fprintf(env.Stdout, "targets:     %s\n", strings.Join(reg.SinkDrivers(), ", "))
	fmt.Fprintf(env.Stdout, "write modes: %s\n", strings.Join(modes, ", "))
	fmt.Fprintf(env.Stdout, "transforms:  %s\n", strings.Join(transform.NewRegistry().Names(), ", "))
	return ExitOK
}

// --- version ---------------------------------------------------------

type versionCmd struct{}

func (versionCmd) Name() string     { return "version" }
func (versionCmd) Synopsis() string { return "print build information" }

func (versionCmd) Run(_ context.Context, env Env, _ []string) int {
	fmt.Fprintln(env.Stdout, version.String())
	return ExitOK
}
