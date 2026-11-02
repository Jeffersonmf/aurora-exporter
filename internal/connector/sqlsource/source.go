// Package sqlsource implements core.Source on top of database/sql, so
// any database with a Go driver (Redshift and PostgreSQL through
// lib/pq, MySQL through go-sql-driver/mysql, ...) can be read from
// without new code.
package sqlsource

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// BinaryTypes reports whether a column of the given database type
// holds genuinely binary data. Drivers hand back many textual types
// (NUMERIC, DECIMAL, JSON, ...) as []byte; those are converted to
// string so sinks bind them as text instead of as binary blobs, which
// PostgreSQL would otherwise store as bytea escape sequences.
type BinaryTypes func(databaseType string) bool

// PostgresBinaryTypes covers PostgreSQL and Redshift.
func PostgresBinaryTypes(t string) bool {
	switch strings.ToUpper(t) {
	case "BYTEA", "VARBYTE":
		return true
	}
	return false
}

// MySQLBinaryTypes covers MySQL and Aurora MySQL.
func MySQLBinaryTypes(t string) bool {
	switch strings.ToUpper(t) {
	case "BINARY", "VARBINARY", "BLOB", "TINYBLOB", "MEDIUMBLOB", "LONGBLOB", "BIT", "GEOMETRY":
		return true
	}
	return false
}

// Source reads query results from a *sql.DB.
type Source struct {
	db       *sql.DB
	isBinary BinaryTypes
}

// New returns a Source over db. A nil isBinary treats every []byte
// value as text.
func New(db *sql.DB, isBinary BinaryTypes) *Source {
	if isBinary == nil {
		isBinary = func(string) bool { return false }
	}
	return &Source{db: db, isBinary: isBinary}
}

// Open runs q and returns a cursor over its result.
func (s *Source) Open(ctx context.Context, q core.Query) (core.Cursor, error) {
	if q.BatchSize <= 0 {
		return nil, fmt.Errorf("batch size must be positive, got %d", q.BatchSize)
	}

	// The cursor owns this context: Close cancels it before closing
	// the rows, so abandoning a huge result set (a failed write, a
	// shutdown) asks the server to stop sending instead of draining
	// every remaining row over the wire.
	ctx, cancel := context.WithCancel(ctx)
	rows, err := s.db.QueryContext(ctx, q.SQL)
	if err != nil {
		cancel()
		return nil, err
	}

	types, err := rows.ColumnTypes()
	if err != nil {
		cancel()
		return nil, errors.Join(err, rows.Close())
	}

	cur := &cursor{
		rows:      rows,
		cancel:    cancel,
		batchSize: q.BatchSize,
		columns:   make([]core.Column, len(types)),
		binary:    make([]bool, len(types)),
	}
	for i, t := range types {
		cur.columns[i] = core.Column{Name: t.Name(), DatabaseType: t.DatabaseTypeName()}
		cur.binary[i] = s.isBinary(t.DatabaseTypeName())
	}
	return cur, nil
}

// Ping implements core.Pinger.
func (s *Source) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

// Close releases the connection pool.
func (s *Source) Close() error { return s.db.Close() }

type cursor struct {
	rows      *sql.Rows
	cancel    context.CancelFunc
	batchSize int
	columns   []core.Column
	binary    []bool
}

func (c *cursor) Columns() []core.Column { return c.columns }

func (c *cursor) Next(ctx context.Context) (core.Batch, error) {
	out := make([]core.Row, 0, c.batchSize)
	for len(out) < c.batchSize {
		if err := ctx.Err(); err != nil {
			return core.Batch{}, err
		}
		if !c.rows.Next() {
			break
		}
		row, err := c.scan()
		if err != nil {
			return core.Batch{}, err
		}
		out = append(out, row)
	}
	if err := c.rows.Err(); err != nil {
		return core.Batch{}, err
	}
	if len(out) == 0 {
		return core.Batch{}, io.EOF
	}
	return core.Batch{Columns: c.columns, Rows: out}, nil
}

func (c *cursor) scan() (core.Row, error) {
	values := make(core.Row, len(c.columns))
	ptrs := make([]any, len(values))
	for i := range values {
		ptrs[i] = &values[i]
	}
	// Scanning into *any copies []byte values, so the row stays valid
	// after the driver reuses its buffers.
	if err := c.rows.Scan(ptrs...); err != nil {
		return nil, err
	}
	for i, v := range values {
		if b, ok := v.([]byte); ok && !c.binary[i] {
			values[i] = string(b)
		}
	}
	return values, nil
}

func (c *cursor) Close() error {
	c.cancel()
	return c.rows.Close()
}
