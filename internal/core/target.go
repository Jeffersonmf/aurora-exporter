package core

import "fmt"

// WriteMode selects how a Sink treats rows already present in the
// target table.
type WriteMode string

const (
	// ModeAppend inserts rows and leaves existing ones untouched.
	ModeAppend WriteMode = "append"
	// ModeReplace deletes every existing row, then inserts, in one
	// transaction. Readers keep seeing the previous data until commit
	// (MVCC), at the cost of a slower DELETE. This is the safe default
	// for full snapshots.
	ModeReplace WriteMode = "replace"
	// ModeTruncate is ModeReplace with TRUNCATE instead of DELETE:
	// faster and bloat-free, but on PostgreSQL it locks the table
	// against readers until commit, and on MySQL it is not
	// transactional at all (it commits implicitly).
	ModeTruncate WriteMode = "truncate"
	// ModeUpsert inserts new rows and updates existing ones, matched
	// on Target.Keys.
	ModeUpsert WriteMode = "upsert"
)

// WriteModes lists every supported mode, in documentation order.
func WriteModes() []WriteMode {
	return []WriteMode{ModeAppend, ModeReplace, ModeTruncate, ModeUpsert}
}

// Validate reports whether m is a known mode.
func (m WriteMode) Validate() error {
	for _, known := range WriteModes() {
		if m == known {
			return nil
		}
	}
	return fmt.Errorf("unknown write mode %q (valid: %v)", m, WriteModes())
}

// Target is where a pipeline writes.
type Target struct {
	// Table is the destination table, optionally schema-qualified
	// ("schema.table").
	Table string
	Mode  WriteMode
	// Keys are the columns identifying a row, required by ModeUpsert.
	Keys []string
}
