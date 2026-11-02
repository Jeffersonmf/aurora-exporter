package sqlsink

import (
	"errors"
	"fmt"
	"strings"
)

// Dialect captures everything that differs between SQL databases for
// the statements this package generates. Supporting a new database
// means implementing Dialect, not touching the Sink or the strategies
// (Open/Closed); the Sink only ever talks to this interface
// (Dependency Inversion).
type Dialect interface {
	Name() string
	// QuoteIdent quotes a single identifier (column or table name).
	QuoteIdent(ident string) string
	// Placeholder returns the bind parameter for the n-th argument
	// (1-based) of a statement.
	Placeholder(n int) string
	// MaxParams is the most bind parameters one statement may carry.
	MaxParams() int
	// DeleteAllSQL removes every row of table, transactionally.
	DeleteAllSQL(table string) string
	// TruncateSQL empties table, or reports why the dialect cannot do
	// so inside a transaction.
	TruncateSQL(table string) (string, error)
	// UpsertSuffix is appended to an INSERT to make it update rows
	// that already exist, matched on keys.
	UpsertSuffix(columns, keys []string) string
}

// QuoteTable quotes a possibly schema-qualified table name
// ("schema.table") part by part.
func QuoteTable(d Dialect, table string) string {
	parts := strings.Split(table, ".")
	for i, p := range parts {
		parts[i] = d.QuoteIdent(p)
	}
	return strings.Join(parts, ".")
}

// Postgres is the dialect for PostgreSQL and Aurora PostgreSQL.
type Postgres struct{}

// Name implements Dialect.
func (Postgres) Name() string { return "postgres" }

// QuoteIdent implements Dialect.
func (Postgres) QuoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// Placeholder implements Dialect.
func (Postgres) Placeholder(n int) string { return fmt.Sprintf("$%d", n) }

// MaxParams implements Dialect: the wire protocol counts parameters in
// an int16.
func (Postgres) MaxParams() int { return 65535 }

// DeleteAllSQL implements Dialect.
func (d Postgres) DeleteAllSQL(table string) string {
	return "DELETE FROM " + QuoteTable(d, table)
}

// TruncateSQL implements Dialect. TRUNCATE is transactional on
// PostgreSQL, but holds an ACCESS EXCLUSIVE lock until commit.
func (d Postgres) TruncateSQL(table string) (string, error) {
	return "TRUNCATE TABLE " + QuoteTable(d, table), nil
}

// UpsertSuffix implements Dialect.
func (d Postgres) UpsertSuffix(columns, keys []string) string {
	conflict := " ON CONFLICT (" + quoteList(d, keys) + ")"
	updates := nonKeys(columns, keys)
	if len(updates) == 0 {
		return conflict + " DO NOTHING"
	}
	sets := make([]string, len(updates))
	for i, c := range updates {
		q := d.QuoteIdent(c)
		sets[i] = q + " = EXCLUDED." + q
	}
	return conflict + " DO UPDATE SET " + strings.Join(sets, ", ")
}

// MySQL is the dialect for MySQL and Aurora MySQL.
type MySQL struct{}

// Name implements Dialect.
func (MySQL) Name() string { return "mysql" }

// QuoteIdent implements Dialect.
func (MySQL) QuoteIdent(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

// Placeholder implements Dialect.
func (MySQL) Placeholder(int) string { return "?" }

// MaxParams implements Dialect.
func (MySQL) MaxParams() int { return 65535 }

// DeleteAllSQL implements Dialect.
func (d MySQL) DeleteAllSQL(table string) string {
	return "DELETE FROM " + QuoteTable(d, table)
}

// ErrTruncateNotTransactional is returned by MySQL.TruncateSQL.
var ErrTruncateNotTransactional = errors.New(
	"mysql commits implicitly on TRUNCATE, so a failed run could not be rolled back; use mode \"replace\"")

// TruncateSQL implements Dialect. It always fails: on MySQL, TRUNCATE
// ends the current transaction, which would leave every row inserted
// after it outside of any rollback.
func (MySQL) TruncateSQL(string) (string, error) {
	return "", ErrTruncateNotTransactional
}

// UpsertSuffix implements Dialect. MySQL matches on any unique index
// of the table; keys only decide which columns are left untouched.
// VALUES() is used instead of the 8.0.19+ row alias syntax so the
// statement also runs on Aurora MySQL 2 (MySQL 5.7 compatible).
func (d MySQL) UpsertSuffix(columns, keys []string) string {
	updates := nonKeys(columns, keys)
	if len(updates) == 0 {
		q := d.QuoteIdent(keys[0])
		return " ON DUPLICATE KEY UPDATE " + q + " = " + q
	}
	sets := make([]string, len(updates))
	for i, c := range updates {
		q := d.QuoteIdent(c)
		sets[i] = q + " = VALUES(" + q + ")"
	}
	return " ON DUPLICATE KEY UPDATE " + strings.Join(sets, ", ")
}

func quoteList(d Dialect, idents []string) string {
	quoted := make([]string, len(idents))
	for i, id := range idents {
		quoted[i] = d.QuoteIdent(id)
	}
	return strings.Join(quoted, ", ")
}

func nonKeys(columns, keys []string) []string {
	isKey := make(map[string]bool, len(keys))
	for _, k := range keys {
		isKey[k] = true
	}
	out := make([]string, 0, len(columns))
	for _, c := range columns {
		if !isKey[c] {
			out = append(out, c)
		}
	}
	return out
}
