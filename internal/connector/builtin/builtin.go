// Package builtin registers every connector shipped with
// aurora-exporter. It is the only package that imports concrete
// database drivers, so the rest of the application stays driver-free.
package builtin

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq" // registers the "postgres" database/sql driver

	"github.com/Jeffersonmf/aurora-exporter/internal/connector/jsonl"
	"github.com/Jeffersonmf/aurora-exporter/internal/connector/sqlsink"
	"github.com/Jeffersonmf/aurora-exporter/internal/connector/sqlsource"
	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/registry"
)

// Driver names accepted in configuration.
const (
	Redshift         = "redshift"
	Postgres         = "postgres"
	AuroraPostgreSQL = "aurora-postgresql"
	MySQL            = "mysql"
	AuroraMySQL      = "aurora-mysql"
	JSONL            = "jsonl"
)

// Register adds every built-in driver to r. stdout is where the jsonl
// sink writes when its path is "-".
func Register(r *registry.Registry, stdout io.Writer) {
	for _, name := range []string{Redshift, Postgres, AuroraPostgreSQL} {
		r.RegisterSource(name, postgresSource)
	}
	for _, name := range []string{Postgres, AuroraPostgreSQL} {
		r.RegisterSink(name, sqlSink("postgres", sqlsink.Postgres{}, normalizePostgresDSN))
	}
	for _, name := range []string{MySQL, AuroraMySQL} {
		r.RegisterSource(name, mysqlSource)
		r.RegisterSink(name, sqlSink("mysql", sqlsink.MySQL{}, normalizeMySQLDSN))
	}
	r.RegisterSink(JSONL, func(c registry.Connection) (core.Sink, error) {
		path := c.Options["path"]
		if path == "" {
			path = jsonl.Stdout
		}
		return jsonl.New(path, stdout), nil
	})
}

func postgresSource(c registry.Connection) (core.Source, error) {
	dsn, err := normalizePostgresDSN(c.DSN)
	if err != nil {
		return nil, err
	}
	db, err := open("postgres", dsn, c)
	if err != nil {
		return nil, err
	}
	return sqlsource.New(db, sqlsource.PostgresBinaryTypes), nil
}

func mysqlSource(c registry.Connection) (core.Source, error) {
	dsn, err := normalizeMySQLDSN(c.DSN)
	if err != nil {
		return nil, err
	}
	db, err := open("mysql", dsn, c)
	if err != nil {
		return nil, err
	}
	return sqlsource.New(db, sqlsource.MySQLBinaryTypes), nil
}

// dsnNormalizer rewrites a user-supplied DSN into what the driver
// expects.
type dsnNormalizer func(dsn string) (string, error)

func sqlSink(driver string, d sqlsink.Dialect, normalize dsnNormalizer) registry.SinkFactory {
	return func(c registry.Connection) (core.Sink, error) {
		dsn, err := normalize(c.DSN)
		if err != nil {
			return nil, err
		}
		rows, err := c.IntOption("rows_per_statement", 0)
		if err != nil {
			return nil, err
		}
		db, err := open(driver, dsn, c)
		if err != nil {
			return nil, err
		}
		return sqlsink.New(db, d, sqlsink.WithRowsPerStatement(rows)), nil
	}
}

func open(driver, dsn string, c registry.Connection) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("connection %q: dsn is empty", c.Name)
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if c.Pool.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.Pool.MaxOpenConns)
	}
	if c.Pool.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.Pool.MaxIdleConns)
	}
	if c.Pool.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(c.Pool.ConnMaxLifetime)
	}
	if c.Pool.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(c.Pool.ConnMaxIdleTime)
	}
	return db, nil
}

// normalizePostgresDSN accepts redshift:// URLs (a common way to write
// Redshift endpoints) by rewriting them to the postgres:// scheme
// lib/pq understands: Redshift speaks the PostgreSQL wire protocol.
func normalizePostgresDSN(dsn string) (string, error) {
	if rest, ok := strings.CutPrefix(dsn, "redshift://"); ok {
		return "postgres://" + rest, nil
	}
	return dsn, nil
}

// normalizeMySQLDSN forces parseTime=true, so DATE/DATETIME columns
// arrive as time.Time instead of raw bytes.
func normalizeMySQLDSN(dsn string) (string, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	cfg.ParseTime = true
	return cfg.FormatDSN(), nil
}
