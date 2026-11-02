// Package cli implements the aurora-exporter command line. Each
// subcommand is a Command; the dispatcher only knows the interface, so
// a new subcommand is a new type plus one entry in commands().
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Jeffersonmf/aurora-exporter/internal/app"
	"github.com/Jeffersonmf/aurora-exporter/internal/config"
	"github.com/Jeffersonmf/aurora-exporter/internal/connector/builtin"
	"github.com/Jeffersonmf/aurora-exporter/internal/logging"
	"github.com/Jeffersonmf/aurora-exporter/internal/registry"
	"github.com/Jeffersonmf/aurora-exporter/internal/transform"
)

// Exit codes.
const (
	ExitOK      = 0
	ExitFailure = 1 // a run failed, the server stopped with an error
	ExitUsage   = 2 // bad flags, invalid configuration
)

// ConfigEnv names the environment variable holding the default config
// path.
const ConfigEnv = "AURORA_EXPORTER_CONFIG"

const defaultConfigPath = "config/aurora-exporter.yaml"

// Env is everything a command touches outside its arguments, so tests
// can run commands in-process with fake connectors and captured output.
type Env struct {
	Stdout io.Writer
	// Stderr receives logs and diagnostics. Logs never go to Stdout,
	// which stays clean for command output (and the jsonl "-" sink).
	Stderr io.Writer
	Getenv func(string) string
	// Connectors builds the driver registry; defaults to the built-ins.
	Connectors func(stdout io.Writer) *registry.Registry
}

// Command is one subcommand.
type Command interface {
	Name() string
	Synopsis() string
	Run(ctx context.Context, env Env, args []string) int
}

func commands() []Command {
	return []Command{serveCmd{}, runCmd{}, validateCmd{}, driversCmd{}, versionCmd{}}
}

// Main runs the command line with the real process environment.
func Main(ctx context.Context, args []string) int {
	return Run(ctx, Env{Stdout: os.Stdout, Stderr: os.Stderr, Getenv: os.Getenv}, args)
}

// Run dispatches args (without the program name) to a command.
func Run(ctx context.Context, env Env, args []string) int {
	if env.Connectors == nil {
		env.Connectors = DefaultConnectors
	}
	if env.Getenv == nil {
		env.Getenv = func(string) string { return "" }
	}

	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		usage(env.Stdout)
		if len(args) == 0 {
			return ExitUsage
		}
		return ExitOK
	}
	for _, c := range commands() {
		if c.Name() == args[0] {
			return c.Run(ctx, env, args[1:])
		}
	}
	fmt.Fprintf(env.Stderr, "unknown command %q\n\n", args[0])
	usage(env.Stderr)
	return ExitUsage
}

// DefaultConnectors is a registry with every built-in driver.
func DefaultConnectors(stdout io.Writer) *registry.Registry {
	r := registry.New()
	builtin.Register(r, stdout)
	return r
}

func usage(w io.Writer) {
	fmt.Fprint(w, "aurora-exporter: export query results from Redshift (or any SQL source) to any supported database.\n\n")
	fmt.Fprint(w, "Usage:\n  aurora-exporter <command> [flags]\n\nCommands:\n")
	for _, c := range commands() {
		fmt.Fprintf(w, "  %-10s %s\n", c.Name(), c.Synopsis())
	}
	fmt.Fprintf(w, "\nRun 'aurora-exporter <command> -h' for a command's flags.\n")
	fmt.Fprintf(w, "The config file defaults to $%s, then %s.\n", ConfigEnv, defaultConfigPath)
}

func newFlagSet(env Env, name, synopsis string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(env.Stderr)
	fs.Usage = func() {
		fmt.Fprintf(env.Stderr, "%s\n\nUsage of %s:\n", synopsis, name)
		fs.PrintDefaults()
	}
	return fs
}

func configFlag(fs *flag.FlagSet, env Env) *string {
	def := env.Getenv(ConfigEnv)
	if def == "" {
		def = defaultConfigPath
	}
	return fs.String("config", def, "path to the YAML configuration file")
}

// parseFlags returns an exit code when the command must stop: help was
// requested (ExitOK) or the flags were invalid (ExitUsage).
func parseFlags(fs *flag.FlagSet, args []string) (int, bool) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK, true
		}
		return ExitUsage, true
	}
	return 0, false
}

// bootstrap loads the configuration and wires the application.
func bootstrap(env Env, path string) (*app.App, *slog.Logger, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, nil, err
	}
	log, err := logging.New(env.Stderr, cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		return nil, nil, err
	}
	a, err := app.Build(cfg, app.Options{
		Connectors: env.Connectors(env.Stdout),
		Transforms: transform.NewRegistry(),
		Logger:     log,
	})
	if err != nil {
		return nil, nil, err
	}
	return a, log, nil
}
