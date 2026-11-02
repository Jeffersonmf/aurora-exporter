package sqlsink_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/connector/sqlsink"
	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

type exec struct {
	query string
	args  []any
}

type fakeTx struct {
	execs      []exec
	failOn     string
	committed  bool
	rolledBack bool
}

func (f *fakeTx) ExecContext(_ context.Context, q string, args ...any) (sql.Result, error) {
	if f.failOn != "" && strings.Contains(q, f.failOn) {
		return nil, errors.New("exec failed: " + f.failOn)
	}
	f.execs = append(f.execs, exec{query: q, args: args})
	return nil, nil
}

func (f *fakeTx) Commit() error   { f.committed = true; return nil }
func (f *fakeTx) Rollback() error { f.rolledBack = true; return nil }

type fakeBeginner struct{ tx *fakeTx }

func (b *fakeBeginner) BeginTx(context.Context) (sqlsink.Tx, error) { return b.tx, nil }

var columns = []core.Column{{Name: "id"}, {Name: "name"}}

func open(t *testing.T, d sqlsink.Dialect, target core.Target, opts ...sqlsink.Option) (*fakeTx, core.Writer) {
	t.Helper()
	tx := &fakeTx{}
	w, err := sqlsink.NewWithBeginner(&fakeBeginner{tx: tx}, d, opts...).Open(context.Background(), target, columns)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return tx, w
}

func rows(n int) core.Batch {
	b := core.Batch{Columns: columns}
	for i := range n {
		b.Rows = append(b.Rows, core.Row{i, "x"})
	}
	return b
}

func TestSink_PrepareStatementPerMode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		mode core.WriteMode
		want string
	}{
		{core.ModeAppend, ""},
		{core.ModeReplace, `DELETE FROM "analytics"."orders"`},
		{core.ModeTruncate, `TRUNCATE TABLE "analytics"."orders"`},
		{core.ModeUpsert, ""},
	}
	for _, tc := range cases {
		t.Run(string(tc.mode), func(t *testing.T) {
			t.Parallel()
			tx, _ := open(t, sqlsink.Postgres{}, core.Target{Table: "analytics.orders", Mode: tc.mode, Keys: []string{"id"}})
			if tc.want == "" {
				if len(tx.execs) != 0 {
					t.Fatalf("unexpected prepare statement %q", tx.execs[0].query)
				}
				return
			}
			if len(tx.execs) != 1 || tx.execs[0].query != tc.want {
				t.Fatalf("execs = %+v, want %q", tx.execs, tc.want)
			}
		})
	}
}

func TestSink_PostgresInsertAndUpsertSQL(t *testing.T) {
	t.Parallel()
	tx, w := open(t, sqlsink.Postgres{}, core.Target{Table: "t", Mode: core.ModeUpsert, Keys: []string{"id"}})

	if _, err := w.Write(context.Background(), rows(2)); err != nil {
		t.Fatal(err)
	}
	want := `INSERT INTO "t" ("id", "name") VALUES ($1, $2), ($3, $4) ON CONFLICT ("id") DO UPDATE SET "name" = EXCLUDED."name"`
	if got := tx.execs[0].query; got != want {
		t.Fatalf("query =\n%s\nwant\n%s", got, want)
	}
	if len(tx.execs[0].args) != 4 {
		t.Fatalf("args = %v, want 4 flattened values", tx.execs[0].args)
	}
}

func TestSink_MySQLInsertAndUpsertSQL(t *testing.T) {
	t.Parallel()
	tx, w := open(t, sqlsink.MySQL{}, core.Target{Table: "db.t", Mode: core.ModeUpsert, Keys: []string{"id"}})

	if _, err := w.Write(context.Background(), rows(1)); err != nil {
		t.Fatal(err)
	}
	want := "INSERT INTO `db`.`t` (`id`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `name` = VALUES(`name`)"
	if got := tx.execs[0].query; got != want {
		t.Fatalf("query =\n%s\nwant\n%s", got, want)
	}
}

func TestSink_UpsertWithOnlyKeyColumns(t *testing.T) {
	t.Parallel()
	pg := sqlsink.Postgres{}.UpsertSuffix([]string{"id"}, []string{"id"})
	if pg != ` ON CONFLICT ("id") DO NOTHING` {
		t.Fatalf("postgres suffix = %q", pg)
	}
	my := sqlsink.MySQL{}.UpsertSuffix([]string{"id"}, []string{"id"})
	if my != " ON DUPLICATE KEY UPDATE `id` = `id`" {
		t.Fatalf("mysql suffix = %q", my)
	}
}

func TestSink_ChunksRowsPerStatement(t *testing.T) {
	t.Parallel()
	tx, w := open(t, sqlsink.Postgres{}, core.Target{Table: "t", Mode: core.ModeAppend}, sqlsink.WithRowsPerStatement(4))

	n, err := w.Write(context.Background(), rows(10))
	if err != nil {
		t.Fatal(err)
	}
	if n != 10 {
		t.Fatalf("written = %d, want 10", n)
	}
	if len(tx.execs) != 3 {
		t.Fatalf("statements = %d, want 3 (4+4+2)", len(tx.execs))
	}
	if got := len(tx.execs[2].args); got != 4 {
		t.Fatalf("last chunk args = %d, want 4 (2 rows x 2 columns)", got)
	}
}

// One INSERT may never exceed the dialect's bind parameter limit, no
// matter how many rows per statement were configured.
func TestSink_RespectsDialectParameterLimit(t *testing.T) {
	t.Parallel()
	tx, w := open(t, tinyDialect{}, core.Target{Table: "t", Mode: core.ModeAppend}, sqlsink.WithRowsPerStatement(1000))

	if _, err := w.Write(context.Background(), rows(5)); err != nil {
		t.Fatal(err)
	}
	for _, e := range tx.execs {
		if len(e.args) > (tinyDialect{}).MaxParams() {
			t.Fatalf("statement carried %d params, limit is %d", len(e.args), (tinyDialect{}).MaxParams())
		}
	}
}

type tinyDialect struct{ sqlsink.Postgres }

func (tinyDialect) MaxParams() int { return 4 }

func TestSink_RejectsMisshapenRows(t *testing.T) {
	t.Parallel()
	_, w := open(t, sqlsink.Postgres{}, core.Target{Table: "t", Mode: core.ModeAppend})
	b := core.Batch{Columns: columns, Rows: []core.Row{{1}}}
	if _, err := w.Write(context.Background(), b); err == nil {
		t.Fatal("expected an error for a row with the wrong number of values")
	}
}

func TestSink_CommitAndRollbackAreIdempotentlySafe(t *testing.T) {
	t.Parallel()
	tx, w := open(t, sqlsink.Postgres{}, core.Target{Table: "t", Mode: core.ModeAppend})

	if err := w.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := w.Rollback(context.Background()); err != nil {
		t.Fatalf("rollback after commit must be a no-op, got %v", err)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("committed=%v rolledBack=%v, want true/false", tx.committed, tx.rolledBack)
	}
}

func TestSink_OpenValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tx := &fakeTx{}
	s := sqlsink.NewWithBeginner(&fakeBeginner{tx: tx}, sqlsink.Postgres{})
	if _, err := s.Open(ctx, core.Target{Table: "t", Mode: core.ModeUpsert, Keys: []string{"missing"}}, columns); err == nil {
		t.Fatal("expected an error for a key column absent from the result set")
	}
	if _, err := s.Open(ctx, core.Target{Table: "t", Mode: "merge"}, columns); err == nil {
		t.Fatal("expected an error for an unsupported mode")
	}
	if _, err := s.Open(ctx, core.Target{Table: "t", Mode: core.ModeAppend}, nil); err == nil {
		t.Fatal("expected an error for a column-less result")
	}
}

func TestSink_MySQLRefusesNonTransactionalTruncate(t *testing.T) {
	t.Parallel()
	tx := &fakeTx{}
	s := sqlsink.NewWithBeginner(&fakeBeginner{tx: tx}, sqlsink.MySQL{})

	_, err := s.Open(context.Background(), core.Target{Table: "t", Mode: core.ModeTruncate}, columns)
	if !errors.Is(err, sqlsink.ErrTruncateNotTransactional) {
		t.Fatalf("err = %v, want ErrTruncateNotTransactional", err)
	}
	if !tx.rolledBack {
		t.Fatal("transaction must be rolled back when prepare fails")
	}
}

func TestSink_PrepareFailureRollsBack(t *testing.T) {
	t.Parallel()
	tx := &fakeTx{failOn: "DELETE"}
	s := sqlsink.NewWithBeginner(&fakeBeginner{tx: tx}, sqlsink.Postgres{})

	if _, err := s.Open(context.Background(), core.Target{Table: "t", Mode: core.ModeReplace}, columns); err == nil {
		t.Fatal("expected prepare error")
	}
	if !tx.rolledBack {
		t.Fatal("transaction must be rolled back when prepare fails")
	}
}

func TestSink_CustomStrategy(t *testing.T) {
	t.Parallel()
	const mode core.WriteMode = "append_ignore"
	tx, w := open(t, sqlsink.Postgres{}, core.Target{Table: "t", Mode: mode}, sqlsink.WithStrategy(mode, ignoreStrategy{}))

	if _, err := w.Write(context.Background(), rows(1)); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(tx.execs[0].query, " ON CONFLICT DO NOTHING") {
		t.Fatalf("custom strategy suffix missing: %q", tx.execs[0].query)
	}
}

type ignoreStrategy struct{ sqlsink.AppendStrategy }

func (ignoreStrategy) InsertSuffix(sqlsink.Dialect, core.Target, []string) string {
	return " ON CONFLICT DO NOTHING"
}

func TestQuoteIdent_EscapesQuotes(t *testing.T) {
	t.Parallel()
	if got := (sqlsink.Postgres{}).QuoteIdent(`we"ird`); got != `"we""ird"` {
		t.Fatalf("postgres = %s", got)
	}
	if got := (sqlsink.MySQL{}).QuoteIdent("we`ird"); got != "`we``ird`" {
		t.Fatalf("mysql = %s", got)
	}
}
