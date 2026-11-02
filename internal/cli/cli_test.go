package cli_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/cli"
	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/core/coretest"
	"github.com/Jeffersonmf/aurora-exporter/internal/registry"
)

const cfg = `
version: 1
logging: {level: error}
connections:
  src:    {driver: mem}
  broken: {driver: broken}
  out:    {driver: jsonl}
pipelines:
  - name: users
    source: {connection: src, query: "select id, name from users"}
    target: {connection: out, table: users}
  - name: orders
    enabled: false
    source: {connection: src, query: "select 1"}
    target: {connection: out, table: orders}
  - name: failing
    enabled: false
    source: {connection: broken, query: "select 1"}
    target: {connection: out, table: failing}
`

// connectors are the built-ins plus a fake "mem" source with two rows
// and a "broken" source that always fails to open.
func connectors(stdout io.Writer) *registry.Registry {
	r := cli.DefaultConnectors(stdout)
	r.RegisterSource("mem", func(registry.Connection) (core.Source, error) {
		return &coretest.Source{Columns: coretest.Columns(), Rows: coretest.Rows(2)}, nil
	})
	r.RegisterSource("broken", func(registry.Connection) (core.Source, error) {
		return &coretest.Source{OpenErr: errors.New("connection refused")}, nil
	})
	return r
}

func run(t *testing.T, doc string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	env := cli.Env{
		Stdout:     &out,
		Stderr:     &errOut,
		Getenv:     func(k string) string { return map[string]string{cli.ConfigEnv: path}[k] },
		Connectors: connectors,
	}
	code = cli.Run(context.Background(), env, args)
	return code, out.String(), errOut.String()
}

func TestRun_NamedPipelineStreamsToJSONLAndPrintsSummary(t *testing.T) {
	t.Parallel()
	code, out, errOut := run(t, cfg, "run", "users")
	if code != cli.ExitOK {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
	if strings.Count(out, `"name":"row"`) != 2 {
		t.Fatalf("expected 2 JSON lines on stdout, got:\n%s", out)
	}
	if !strings.Contains(out, "users") || !strings.Contains(out, "succeeded") {
		t.Fatalf("summary missing:\n%s", out)
	}
}

func TestRun_AllSkipsDisabledPipelines(t *testing.T) {
	t.Parallel()
	code, out, _ := run(t, cfg, "run", "-all")
	if code != cli.ExitOK {
		t.Fatalf("exit = %d", code)
	}
	if strings.Contains(out, "orders") || strings.Contains(out, "failing") {
		t.Fatalf("disabled pipelines ran:\n%s", out)
	}
}

func TestRun_FailureExitsNonZero(t *testing.T) {
	t.Parallel()
	code, out, _ := run(t, cfg, "run", "failing")
	if code != cli.ExitFailure {
		t.Fatalf("exit = %d, want %d", code, cli.ExitFailure)
	}
	if !strings.Contains(out, "connection refused") {
		t.Fatalf("summary should carry the error:\n%s", out)
	}
}

func TestRun_UsageErrors(t *testing.T) {
	t.Parallel()
	cases := map[string][]string{
		"no pipelines":     {"run"},
		"unknown pipeline": {"run", "nope"},
		"bad flag":         {"run", "-nope"},
		"unknown command":  {"explode"},
		"no command":       {},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if code, _, _ := run(t, cfg, args...); code != cli.ExitUsage {
				t.Fatalf("exit = %d, want %d", code, cli.ExitUsage)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	code, out, _ := run(t, cfg, "validate")
	if code != cli.ExitOK || !strings.Contains(out, "OK (3 pipelines, 0 scheduled, 3 connections)") {
		t.Fatalf("exit = %d, out = %s", code, out)
	}

	code, _, errOut := run(t, strings.Replace(cfg, "driver: mem", "driver: oracle", 1), "validate")
	if code != cli.ExitUsage || !strings.Contains(errOut, `driver "oracle"`) {
		t.Fatalf("exit = %d, stderr = %s", code, errOut)
	}
}

func TestDriversAndVersion(t *testing.T) {
	t.Parallel()
	_, out, _ := run(t, cfg, "drivers")
	for _, want := range []string{"redshift", "aurora-postgresql", "aurora-mysql", "jsonl", "upsert", "sanitize_utf8"} {
		if !strings.Contains(out, want) {
			t.Errorf("drivers output missing %q:\n%s", want, out)
		}
	}

	if code, out, _ := run(t, cfg, "version"); code != cli.ExitOK || !strings.HasPrefix(out, "aurora-exporter dev") {
		t.Fatalf("version: exit = %d, out = %q", code, out)
	}
	if code, out, _ := run(t, cfg, "help"); code != cli.ExitOK || !strings.Contains(out, "Commands:") {
		t.Fatalf("help: exit = %d", code)
	}
}
