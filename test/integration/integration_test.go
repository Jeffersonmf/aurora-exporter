//go:build integration

// Package integration runs real exports between real databases started
// by docker-compose.yml (make test-integration). It goes through the
// same path as production: YAML config -> app.Build with the built-in
// drivers -> service.Execute.
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/Jeffersonmf/aurora-exporter/internal/app"
	"github.com/Jeffersonmf/aurora-exporter/internal/cli"
	"github.com/Jeffersonmf/aurora-exporter/internal/config"
	"github.com/Jeffersonmf/aurora-exporter/internal/run"
	"github.com/Jeffersonmf/aurora-exporter/internal/transform"
)

const seededRows = 2500

var dsn = map[string]string{
	"SOURCE_DSN":   env("IT_SOURCE_DSN", "postgres://test:test@localhost:55432/source?sslmode=disable"),
	"POSTGRES_DSN": env("IT_POSTGRES_DSN", "postgres://test:test@localhost:55433/target?sslmode=disable"),
	"MYSQL_DSN":    env("IT_MYSQL_DSN", "test:test@tcp(localhost:53306)/target"),
}

var source, pg, my *sql.DB

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestMain(m *testing.M) {
	var err error
	source, err = connect("postgres", dsn["SOURCE_DSN"])
	if err == nil {
		pg, err = connect("postgres", dsn["POSTGRES_DSN"])
	}
	if err == nil {
		my, err = connect("mysql", dsn["MYSQL_DSN"]+"?parseTime=true")
	}
	if err == nil {
		err = seed()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration setup failed (is `docker compose up -d` running?): %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func connect(driver, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(90 * time.Second)
	for {
		if err = db.PingContext(context.Background()); err == nil || time.Now().After(deadline) {
			return db, err
		}
		time.Sleep(time.Second)
	}
}

func seed() error {
	stmts := []string{
		`DROP TABLE IF EXISTS events`,
		`CREATE TABLE events (
			id int PRIMARY KEY, name text, amount numeric(12,2), created_at timestamp,
			active boolean, payload bytea, note text)`,
		// Row 1 covers non-ASCII text and binary data; row 2 is all NULLs
		// except the key, which the original exporter silently turned
		// into 0 / false / 0001-01-01.
		`INSERT INTO events VALUES
			(1, 'alpha ção ✓', 10.50, '2024-01-01 10:00:00', true, '\x0102'::bytea, NULL),
			(2, NULL, NULL, NULL, NULL, NULL, NULL)`,
		fmt.Sprintf(`INSERT INTO events
			SELECT g, 'name-' || g, (g %% 1000) + 0.25, timestamp '2024-01-01' + g * interval '1 minute',
			       g %% 2 = 0, NULL, 'generated'
			FROM generate_series(3, %d) g`, seededRows),
	}
	for _, s := range stmts {
		if _, err := source.ExecContext(context.Background(), s); err != nil {
			return fmt.Errorf("%s: %w", s, err)
		}
	}
	return nil
}

func resetPostgres(t *testing.T) {
	t.Helper()
	mustExec(t, pg, `CREATE SCHEMA IF NOT EXISTS hot`)
	mustExec(t, pg, `DROP TABLE IF EXISTS hot.events`)
	mustExec(t, pg, `CREATE TABLE hot.events (
		id int PRIMARY KEY, name text, amount numeric(12,2) CHECK (amount < 100000), created_at timestamp,
		active boolean, payload bytea, note text)`)
}

func resetMySQL(t *testing.T) {
	t.Helper()
	mustExec(t, my, `DROP TABLE IF EXISTS events`)
	mustExec(t, my, `CREATE TABLE events (
		id INT PRIMARY KEY, name VARCHAR(255), amount DECIMAL(12,2), created_at DATETIME,
		active BOOLEAN, payload VARBINARY(16), note TEXT) CHARACTER SET utf8mb4`)
}

func mustExec(t *testing.T, db *sql.DB, q string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), q); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

func scalar[T any](t *testing.T, db *sql.DB, q string) T {
	t.Helper()
	var v T
	if err := db.QueryRowContext(context.Background(), q).Scan(&v); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return v
}

// export runs one pipeline exporting query from the source into table on
// target ("postgres" or "mysql") with mode/keys.
func export(t *testing.T, target, table, mode, query string, keys ...string) (run.Run, error) {
	t.Helper()
	keyList := ""
	if len(keys) > 0 {
		keyList = ", keys: [" + strings.Join(keys, ", ") + "]"
	}
	doc := fmt.Sprintf(`
version: 1
logging: {level: error}
defaults:
  batch_size: 1000
connections:
  redshift: {driver: redshift, dsn: "${SOURCE_DSN}"}
  postgres: {driver: aurora-postgresql, dsn: "${POSTGRES_DSN}", options: {rows_per_statement: "300"}}
  mysql:    {driver: aurora-mysql, dsn: "${MYSQL_DSN}", options: {rows_per_statement: "300"}}
pipelines:
  - name: it
    source: {connection: redshift, query: %q}
    target: {connection: %s, table: %s, mode: %s%s}
`, query, target, table, mode, keyList)

	cfg, err := config.Parse([]byte(doc), func(k string) (string, bool) { v, ok := dsn[k]; return v, ok })
	if err != nil {
		t.Fatal(err)
	}
	a, err := app.Build(cfg, app.Options{Connectors: cli.DefaultConnectors(io.Discard), Transforms: transform.NewRegistry()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()
	return a.Service.Execute(context.Background(), "it", run.TriggerCLI)
}

const all = "SELECT * FROM events"

func TestPostgres_ReplaceIsIdempotentAndPreservesValues(t *testing.T) {
	resetPostgres(t)

	for range 2 {
		r, err := export(t, "postgres", "hot.events", "replace", all)
		if err != nil {
			t.Fatal(err)
		}
		if r.RowsWritten != seededRows || r.Batches < 3 {
			t.Fatalf("run = %+v, want %d rows over several batches", r, seededRows)
		}
	}

	if n := scalar[int](t, pg, `SELECT count(*) FROM hot.events`); n != seededRows {
		t.Fatalf("count = %d after two replace runs, want %d", n, seededRows)
	}
	if got := scalar[string](t, pg, `SELECT name || '|' || amount || '|' || created_at || '|' || active || '|' || encode(payload, 'hex') FROM hot.events WHERE id = 1`); got != "alpha ção ✓|10.50|2024-01-01 10:00:00|true|0102" {
		t.Fatalf("row 1 = %q", got)
	}
	if nulls := scalar[int](t, pg, `SELECT num_nulls(name, amount, created_at, active, payload, note) FROM hot.events WHERE id = 2`); nulls != 6 {
		t.Fatalf("row 2 has %d NULLs, want 6: NULL must stay NULL", nulls)
	}
	if sum := scalar[string](t, pg, `SELECT sum(amount)::text FROM hot.events`); sum != scalar[string](t, source, `SELECT sum(amount)::text FROM events`) {
		t.Fatalf("sum(amount) differs between source and target: %s", sum)
	}
}

func TestPostgres_FailedRunLeavesPreviousDataIntact(t *testing.T) {
	resetPostgres(t)
	if _, err := export(t, "postgres", "hot.events", "replace", all); err != nil {
		t.Fatal(err)
	}

	// Row 2400 violates the target's CHECK constraint, deep into the run
	// (after the DELETE and two committed-looking batches).
	bad := `SELECT id, name, CASE WHEN id = 2400 THEN 999999 ELSE amount END AS amount,
	        created_at, active, payload, 'second run' AS note FROM events ORDER BY id`
	r, err := export(t, "postgres", "hot.events", "replace", bad)
	if err == nil || !strings.Contains(err.Error(), "check constraint") {
		t.Fatalf("err = %v, want the CHECK violation", err)
	}
	if r.Status != run.StatusFailed {
		t.Fatalf("status = %s", r.Status)
	}

	if n := scalar[int](t, pg, `SELECT count(*) FROM hot.events`); n != seededRows {
		t.Fatalf("count = %d, the failed run must not have deleted anything", n)
	}
	if n := scalar[int](t, pg, `SELECT count(*) FROM hot.events WHERE note = 'second run'`); n != 0 {
		t.Fatalf("%d rows from the failed run are visible", n)
	}
}

func TestPostgres_UpsertAppendTruncate(t *testing.T) {
	resetPostgres(t)
	if _, err := export(t, "postgres", "hot.events", "replace", all); err != nil {
		t.Fatal(err)
	}

	upd := `SELECT id, name || '-v2' AS name FROM events WHERE id BETWEEN 3 AND 12
	        UNION ALL SELECT 9001, 'new'`
	if _, err := export(t, "postgres", "hot.events", "upsert", upd, "id"); err != nil {
		t.Fatal(err)
	}
	if n := scalar[int](t, pg, `SELECT count(*) FROM hot.events WHERE name LIKE '%-v2'`); n != 10 {
		t.Fatalf("updated rows = %d, want 10", n)
	}
	if n := scalar[int](t, pg, `SELECT count(*) FROM hot.events`); n != seededRows+1 {
		t.Fatalf("count = %d, want %d (one inserted)", n, seededRows+1)
	}
	if note := scalar[string](t, pg, `SELECT note FROM hot.events WHERE id = 3`); note != "generated" {
		t.Fatalf("upsert overwrote a column it did not select: note = %q", note)
	}

	if _, err := export(t, "postgres", "hot.events", "truncate", `SELECT * FROM events WHERE id <= 100`); err != nil {
		t.Fatal(err)
	}
	if _, err := export(t, "postgres", "hot.events", "append", `SELECT id + 100000 AS id, name FROM events WHERE id <= 5`); err != nil {
		t.Fatal(err)
	}
	if n := scalar[int](t, pg, `SELECT count(*) FROM hot.events`); n != 105 {
		t.Fatalf("count = %d, want 100 (truncate) + 5 (append)", n)
	}
}

func TestMySQL_ReplaceUpsertAndTruncateGuard(t *testing.T) {
	resetMySQL(t)

	for range 2 {
		if _, err := export(t, "mysql", "events", "replace", all); err != nil {
			t.Fatal(err)
		}
	}
	if n := scalar[int](t, my, `SELECT count(*) FROM events`); n != seededRows {
		t.Fatalf("count = %d, want %d", n, seededRows)
	}
	if got := scalar[string](t, my, `SELECT CONCAT_WS('|', name, amount, created_at, active, HEX(payload)) FROM events WHERE id = 1`); got != "alpha ção ✓|10.50|2024-01-01 10:00:00|1|0102" {
		t.Fatalf("row 1 = %q", got)
	}
	if n := scalar[int](t, my, `SELECT count(*) FROM events WHERE id = 2 AND name IS NULL AND amount IS NULL AND created_at IS NULL AND active IS NULL`); n != 1 {
		t.Fatal("row 2 NULLs were not preserved")
	}

	if _, err := export(t, "mysql", "events", "upsert", `SELECT id, name || '-v2' AS name FROM events WHERE id <= 10`, "id"); err != nil {
		t.Fatal(err)
	}
	if n := scalar[int](t, my, `SELECT count(*) FROM events WHERE name LIKE '%-v2'`); n != 9 { // row 2's name is NULL
		t.Fatalf("updated rows = %d, want 9", n)
	}

	_, err := export(t, "mysql", "events", "truncate", all)
	if err == nil || !strings.Contains(err.Error(), "use mode \"replace\"") {
		t.Fatalf("err = %v, want truncate refused on mysql", err)
	}
	if n := scalar[int](t, my, `SELECT count(*) FROM events`); n != seededRows {
		t.Fatalf("count = %d after the refused truncate", n)
	}
}
