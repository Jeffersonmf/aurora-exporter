package sqlsink

import (
	"context"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// Strategy implements one core.WriteMode. Each mode is its own type
// with one reason to change (Single Responsibility), and the Sink
// selects them from a map instead of branching on the mode (Open/Closed:
// a new mode is a new Strategy plus one map entry).
type Strategy interface {
	// Prepare runs once, inside the transaction, before any row.
	Prepare(ctx context.Context, ex Execer, d Dialect, target core.Target) error
	// InsertSuffix is appended to every INSERT statement.
	InsertSuffix(d Dialect, target core.Target, columns []string) string
}

// DefaultStrategies returns the strategy for every core.WriteMode.
func DefaultStrategies() map[core.WriteMode]Strategy {
	return map[core.WriteMode]Strategy{
		core.ModeAppend:   AppendStrategy{},
		core.ModeReplace:  ReplaceStrategy{},
		core.ModeTruncate: TruncateStrategy{},
		core.ModeUpsert:   UpsertStrategy{},
	}
}

// AppendStrategy inserts without touching existing rows.
type AppendStrategy struct{}

// Prepare implements Strategy.
func (AppendStrategy) Prepare(context.Context, Execer, Dialect, core.Target) error { return nil }

// InsertSuffix implements Strategy.
func (AppendStrategy) InsertSuffix(Dialect, core.Target, []string) string { return "" }

// ReplaceStrategy deletes every row first, in the same transaction.
type ReplaceStrategy struct{ AppendStrategy }

// Prepare implements Strategy.
func (ReplaceStrategy) Prepare(ctx context.Context, ex Execer, d Dialect, t core.Target) error {
	_, err := ex.ExecContext(ctx, d.DeleteAllSQL(t.Table))
	return err
}

// TruncateStrategy truncates first, in the same transaction, where the
// dialect allows it.
type TruncateStrategy struct{ AppendStrategy }

// Prepare implements Strategy.
func (TruncateStrategy) Prepare(ctx context.Context, ex Execer, d Dialect, t core.Target) error {
	stmt, err := d.TruncateSQL(t.Table)
	if err != nil {
		return err
	}
	_, err = ex.ExecContext(ctx, stmt)
	return err
}

// UpsertStrategy inserts or updates, matched on the target's keys.
type UpsertStrategy struct{}

// Prepare implements Strategy.
func (UpsertStrategy) Prepare(context.Context, Execer, Dialect, core.Target) error { return nil }

// InsertSuffix implements Strategy.
func (UpsertStrategy) InsertSuffix(d Dialect, t core.Target, columns []string) string {
	return d.UpsertSuffix(columns, t.Keys)
}
