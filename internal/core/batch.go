package core

// Column describes one column of a result set.
type Column struct {
	// Name is the column name exactly as the source reported it.
	Name string
	// DatabaseType is the source's own type name (e.g. "VARCHAR",
	// "NUMERIC", "TIMESTAMP"), informational only: sinks write values,
	// they do not translate types.
	DatabaseType string
}

// Row holds one value per Column, in the same order.
type Row []any

// Batch is a chunk of rows sharing the same columns. It is the unit
// that flows from a Source, through Transformers, into a Sink.
type Batch struct {
	Columns []Column
	Rows    []Row
}

// Len returns how many rows the batch holds.
func (b Batch) Len() int { return len(b.Rows) }

// ColumnNames returns the batch's column names, in order.
func (b Batch) ColumnNames() []string {
	names := make([]string, len(b.Columns))
	for i, c := range b.Columns {
		names[i] = c.Name
	}
	return names
}

// Schema returns a row-less copy of the batch: the same columns, no
// data. Transformers receive a schema batch once, before any data, so
// the Sink can be opened with the final (post-transform) columns.
func (b Batch) Schema() Batch {
	return Batch{Columns: append([]Column(nil), b.Columns...)}
}
