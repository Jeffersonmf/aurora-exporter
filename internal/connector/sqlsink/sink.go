// Package sqlsink implements core.Sink for SQL databases: every run is
// one transaction, rows are written with multi-row INSERTs sized to
// the dialect's parameter limit, and the write mode is a pluggable
// Strategy.
package sqlsink

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// defaultRowsPerStatement caps a single INSERT even when the dialect's
// parameter limit would allow more, keeping statements (and the
// server's parse cost) reasonably sized.
const defaultRowsPerStatement = 1000

// Execer is the subset of *sql.Tx the strategies and the writer need.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Tx is a transaction: an Execer that can be committed or rolled back.
type Tx interface {
	Execer
	Commit() error
	Rollback() error
}

// TxBeginner starts transactions. *sql.DB is adapted to it by New;
// tests substitute a fake that records statements.
type TxBeginner interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type dbBeginner struct{ db *sql.DB }

func (b dbBeginner) BeginTx(ctx context.Context) (Tx, error) { return b.db.BeginTx(ctx, nil) }

// Option configures a Sink.
type Option func(*Sink)

// WithRowsPerStatement caps how many rows one INSERT carries.
func WithRowsPerStatement(n int) Option {
	return func(s *Sink) {
		if n > 0 {
			s.rowsPerStatement = n
		}
	}
}

// WithStrategy registers (or overrides) the strategy for mode.
func WithStrategy(mode core.WriteMode, st Strategy) Option {
	return func(s *Sink) { s.strategies[mode] = st }
}

// Sink writes to a SQL database.
type Sink struct {
	beginner         TxBeginner
	dialect          Dialect
	strategies       map[core.WriteMode]Strategy
	rowsPerStatement int
	db               *sql.DB
}

// New returns a Sink over db speaking dialect.
func New(db *sql.DB, dialect Dialect, opts ...Option) *Sink {
	s := NewWithBeginner(dbBeginner{db: db}, dialect, opts...)
	s.db = db
	return s
}

// NewWithBeginner returns a Sink over any TxBeginner.
func NewWithBeginner(b TxBeginner, dialect Dialect, opts ...Option) *Sink {
	s := &Sink{
		beginner:         b,
		dialect:          dialect,
		strategies:       DefaultStrategies(),
		rowsPerStatement: defaultRowsPerStatement,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Open implements core.Sink: it begins a transaction and runs the
// mode's preparation (DELETE, TRUNCATE, ...) inside it.
func (s *Sink) Open(ctx context.Context, target core.Target, columns []core.Column) (core.Writer, error) {
	strategy, ok := s.strategies[target.Mode]
	if !ok {
		return nil, fmt.Errorf("%s sink does not support write mode %q", s.dialect.Name(), target.Mode)
	}
	if len(columns) == 0 {
		return nil, errors.New("cannot write a result set without columns")
	}

	names := make([]string, len(columns))
	for i, c := range columns {
		names[i] = c.Name
	}
	for _, k := range target.Keys {
		if !slices.Contains(names, k) {
			return nil, fmt.Errorf("key column %q is not in the result set %v", k, names)
		}
	}

	tx, err := s.beginner.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	if err := strategy.Prepare(ctx, tx, s.dialect, target); err != nil {
		return nil, errors.Join(fmt.Errorf("prepare %s: %w", target.Mode, err), tx.Rollback())
	}

	return &writer{
		tx:      tx,
		dialect: s.dialect,
		columns: names,
		prefix:  "INSERT INTO " + QuoteTable(s.dialect, target.Table) + " (" + quoteList(s.dialect, names) + ") VALUES ",
		suffix:  strategy.InsertSuffix(s.dialect, target, names),
		perStmt: min(s.rowsPerStatement, s.dialect.MaxParams()/len(names)),
		stmts:   map[int]string{},
	}, nil
}

// Ping implements core.Pinger.
func (s *Sink) Ping(ctx context.Context) error {
	if s.db == nil {
		return nil
	}
	return s.db.PingContext(ctx)
}

// Close releases the connection pool.
func (s *Sink) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

type writer struct {
	tx      Tx
	dialect Dialect
	columns []string
	prefix  string
	suffix  string
	perStmt int
	stmts   map[int]string
	done    bool
}

func (w *writer) Write(ctx context.Context, b core.Batch) (int64, error) {
	var written int64
	for start := 0; start < len(b.Rows); start += w.perStmt {
		chunk := b.Rows[start:min(start+w.perStmt, len(b.Rows))]

		args := make([]any, 0, len(chunk)*len(w.columns))
		for _, row := range chunk {
			if len(row) != len(w.columns) {
				return written, fmt.Errorf("row has %d values, expected %d columns", len(row), len(w.columns))
			}
			args = append(args, row...)
		}

		if _, err := w.tx.ExecContext(ctx, w.statement(len(chunk)), args...); err != nil {
			return written, err
		}
		written += int64(len(chunk))
	}
	return written, nil
}

// statement builds (and caches by row count) the INSERT for n rows.
// Almost every chunk is full-sized, so in practice this builds two
// statements per run: the full one and the final remainder.
func (w *writer) statement(n int) string {
	if stmt, ok := w.stmts[n]; ok {
		return stmt
	}

	var sb strings.Builder
	sb.WriteString(w.prefix)
	param := 1
	for r := range n {
		if r > 0 {
			sb.WriteString(", ")
		}
		sb.WriteByte('(')
		for c := range w.columns {
			if c > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(w.dialect.Placeholder(param))
			param++
		}
		sb.WriteByte(')')
	}
	sb.WriteString(w.suffix)

	stmt := sb.String()
	w.stmts[n] = stmt
	return stmt
}

func (w *writer) Commit(context.Context) error {
	if w.done {
		return errors.New("writer already finished")
	}
	w.done = true
	return w.tx.Commit()
}

func (w *writer) Rollback(context.Context) error {
	if w.done {
		return nil
	}
	w.done = true
	if err := w.tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		return err
	}
	return nil
}
