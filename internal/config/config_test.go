package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Jeffersonmf/aurora-exporter/internal/config"
)

const valid = `
version: 1
defaults:
  batch_size: 1000
  transforms: [sanitize_utf8]
connections:
  redshift: {driver: redshift, dsn: "${REDSHIFT_DSN}"}
  warehouse: {driver: postgres, dsn: "${TARGET_DSN:-postgres://localhost/app}", max_open_conns: 4}
pipelines:
  - name: customers
    source: {connection: redshift, query: "select * from sales.customers where id > $1"}
    target: {connection: warehouse, table: analytics.customers}
    schedule: {every: 1h}
  - name: orders
    enabled: false
    attempts: 1
    timeout: 10m
    source: {connection: redshift, query: "select * from sales.orders"}
    target: {connection: warehouse, table: analytics.orders, mode: upsert, keys: [id]}
    transforms:
      - lowercase_columns
      - type: rename_columns
        columns: {createdat: created_at}
`

func env(vars map[string]string) config.LookupEnv {
	return func(k string) (string, bool) { v, ok := vars[k]; return v, ok }
}

func parse(t *testing.T, doc string, vars map[string]string) *config.Config {
	t.Helper()
	cfg, err := config.Parse([]byte(doc), env(vars))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return cfg
}

func TestParse_ValidFileWithDefaults(t *testing.T) {
	t.Parallel()
	cfg := parse(t, valid, map[string]string{"REDSHIFT_DSN": "postgres://rs"})

	if cfg.Server.Addr != config.DefaultAddr || cfg.Server.ShutdownTimeout.Std() != config.DefaultShutdownTimeout {
		t.Errorf("server defaults not applied: %+v", cfg.Server)
	}
	if cfg.Logging.Level != "info" || cfg.Logging.Format != "json" {
		t.Errorf("logging defaults not applied: %+v", cfg.Logging)
	}

	customers := cfg.Pipelines[0]
	if customers.BatchSize != 1000 || *customers.Attempts != config.DefaultAttempts || customers.Target.Mode != "replace" {
		t.Errorf("pipeline defaults not applied: batch=%d attempts=%d mode=%s", customers.BatchSize, *customers.Attempts, customers.Target.Mode)
	}
	if customers.Schedule.Every.Std() != time.Hour || !customers.IsEnabled() {
		t.Errorf("schedule = %v enabled=%v", customers.Schedule.Every.Std(), customers.IsEnabled())
	}
	if len(customers.Transforms) != 1 || customers.Transforms[0].Type != "sanitize_utf8" {
		t.Errorf("default transforms not inherited: %+v", customers.Transforms)
	}
	if !strings.Contains(customers.Source.Query, "$1") {
		t.Error("a $1 placeholder in SQL must not be treated as an environment variable")
	}

	orders := cfg.Pipelines[1]
	if orders.IsEnabled() || *orders.Attempts != 1 || orders.Timeout.Std() != 10*time.Minute {
		t.Errorf("overrides not kept: enabled=%v attempts=%d timeout=%v", orders.IsEnabled(), *orders.Attempts, orders.Timeout.Std())
	}
	if len(orders.Transforms) != 2 || orders.Transforms[1].Type != "rename_columns" || orders.Transforms[1].Params == nil {
		t.Errorf("transforms override = %+v", orders.Transforms)
	}
}

func TestParse_ExpandsEnvironment(t *testing.T) {
	t.Parallel()
	cfg := parse(t, valid, map[string]string{"REDSHIFT_DSN": "postgres://u:p#a: ss@rs:5439/db"})

	if got := cfg.Connections["redshift"].DSN; got != "postgres://u:p#a: ss@rs:5439/db" {
		t.Errorf("secret with YAML syntax was mangled: %q", got)
	}
	if got := cfg.Connections["warehouse"].DSN; got != "postgres://localhost/app" {
		t.Errorf("default value not applied: %q", got)
	}
}

func TestParse_ExpandedValuesStayStrings(t *testing.T) {
	t.Parallel()
	doc := strings.Replace(valid, `table: analytics.orders,`, `table: "${TABLE}",`, 1)
	cfg := parse(t, doc, map[string]string{"REDSHIFT_DSN": "x", "TABLE": "123"})
	if cfg.Pipelines[1].Target.Table != "123" {
		t.Fatalf("table = %q", cfg.Pipelines[1].Target.Table)
	}
}

func TestParse_EscapedReferenceIsLiteral(t *testing.T) {
	t.Parallel()
	doc := strings.Replace(valid, "select * from sales.orders", "select '$${NOT_ENV}' as x", 1)
	cfg := parse(t, doc, map[string]string{"REDSHIFT_DSN": "x"})
	if got := cfg.Pipelines[1].Source.Query; got != "select '${NOT_ENV}' as x" {
		t.Fatalf("query = %q", got)
	}
}

func TestParse_ReportsEveryMissingVariable(t *testing.T) {
	t.Parallel()
	doc := valid + "\nserver: {auth_token: \"${API_TOKEN}\"}\n"
	_, err := config.Parse([]byte(doc), env(nil))
	if err == nil || !strings.Contains(err.Error(), "API_TOKEN, REDSHIFT_DSN") {
		t.Fatalf("err = %v, want both missing variables listed", err)
	}
}

func TestParse_RejectsUnknownFields(t *testing.T) {
	t.Parallel()
	doc := strings.Replace(valid, "batch_size: 1000", "bacth_size: 1000", 1)
	_, err := config.Parse([]byte(doc), env(map[string]string{"REDSHIFT_DSN": "x"}))
	if err == nil || !strings.Contains(err.Error(), "bacth_size") {
		t.Fatalf("err = %v, want the misspelled key reported", err)
	}
}

func TestValidate_ReportsAllViolations(t *testing.T) {
	t.Parallel()
	doc := `
version: 2
logging: {level: loud}
connections:
  pg: {driver: ""}
pipelines:
  - name: Bad Name
    source: {connection: nope, query: ""}
    target: {connection: pg, table: "", mode: merge}
    schedule: {every: 500ms}
    attempts: 0
  - name: dup
    source: {connection: pg, query: "select 1"}
    target: {connection: pg, table: t, mode: upsert}
    transforms: [{columns: [a]}]
  - name: dup
    source: {connection: pg, query: "select 1"}
    target: {connection: pg, table: t}
`
	_, err := config.Parse([]byte(doc), env(nil))
	if err == nil {
		t.Fatal("expected validation errors")
	}
	for _, want := range []string{
		"version: must be 1",
		"logging.level",
		"connections.pg.driver",
		"name: must match",
		`unknown connection "nope"`,
		"source.query: is required",
		"target.table: is required",
		`unknown write mode "merge"`,
		"schedule.every: must be at least 1s",
		"attempts: must be at least 1",
		"upsert requires at least one key",
		"transforms[0]: type is required",
		"name: duplicated",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing violation %q in:\n%v", want, err)
		}
	}
}

func TestParse_InvalidDuration(t *testing.T) {
	t.Parallel()
	doc := strings.Replace(valid, "every: 1h", "every: hourly", 1)
	_, err := config.Parse([]byte(doc), env(map[string]string{"REDSHIFT_DSN": "x"}))
	if err == nil || !strings.Contains(err.Error(), "invalid duration") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoad_ExampleConfigsAreValid(t *testing.T) {
	for _, path := range []string{
		"../../config/aurora-exporter.example.yaml",
		"../../examples/quickstart/aurora-exporter.yaml",
	} {
		t.Run(path, func(t *testing.T) {
			t.Setenv("REDSHIFT_DSN", "postgres://example")
			t.Setenv("TARGET_DSN", "postgres://example")
			if _, err := config.Load(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}
