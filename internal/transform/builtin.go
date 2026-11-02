package transform

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// SanitizeUTF8 replaces invalid UTF-8 sequences and strips NUL bytes
// from every string value. PostgreSQL rejects both in text columns
// ("invalid byte sequence for encoding UTF8: 0x00"), and Redshift will
// happily return them, so one bad row would otherwise fail an entire
// run.
func SanitizeUTF8() core.Transformer {
	return core.TransformerFunc(func(b core.Batch) (core.Batch, error) {
		for _, row := range b.Rows {
			for i, v := range row {
				if s, ok := v.(string); ok {
					row[i] = sanitize(s)
				}
			}
		}
		return b, nil
	})
}

func sanitize(s string) string {
	if utf8.ValidString(s) && !strings.ContainsRune(s, 0) {
		return s
	}
	return strings.ReplaceAll(strings.ToValidUTF8(s, ""), "\x00", "")
}

func newSanitizeUTF8(Decoder) (core.Transformer, error) { return SanitizeUTF8(), nil }

// RenameColumns renames columns according to mapping (old -> new).
// Columns not in mapping keep their name.
func RenameColumns(mapping map[string]string) core.Transformer {
	return core.TransformerFunc(func(b core.Batch) (core.Batch, error) {
		cols := make([]core.Column, len(b.Columns))
		for i, c := range b.Columns {
			if renamed, ok := mapping[c.Name]; ok {
				c.Name = renamed
			}
			cols[i] = c
		}
		b.Columns = cols
		return b, nil
	})
}

func newRenameColumns(params Decoder) (core.Transformer, error) {
	var p struct {
		Columns map[string]string `yaml:"columns"`
	}
	if err := decode(params, &p); err != nil {
		return nil, err
	}
	if len(p.Columns) == 0 {
		return nil, errors.New("columns: at least one old -> new mapping is required")
	}
	return RenameColumns(p.Columns), nil
}

// DropColumns removes the named columns and their values.
func DropColumns(names ...string) core.Transformer {
	drop := make(map[string]bool, len(names))
	for _, n := range names {
		drop[n] = true
	}
	return core.TransformerFunc(func(b core.Batch) (core.Batch, error) {
		keep := make([]int, 0, len(b.Columns))
		cols := make([]core.Column, 0, len(b.Columns))
		for i, c := range b.Columns {
			if !drop[c.Name] {
				keep = append(keep, i)
				cols = append(cols, c)
			}
		}
		if len(cols) == 0 {
			return core.Batch{}, errors.New("drop_columns would remove every column")
		}

		rows := make([]core.Row, len(b.Rows))
		for r, row := range b.Rows {
			out := make(core.Row, len(keep))
			for j, i := range keep {
				out[j] = row[i]
			}
			rows[r] = out
		}
		return core.Batch{Columns: cols, Rows: rows}, nil
	})
}

func newDropColumns(params Decoder) (core.Transformer, error) {
	var p struct {
		Columns []string `yaml:"columns"`
	}
	if err := decode(params, &p); err != nil {
		return nil, err
	}
	if len(p.Columns) == 0 {
		return nil, errors.New("columns: at least one column name is required")
	}
	return DropColumns(p.Columns...), nil
}

// LowercaseColumns lowercases every column name, matching how
// PostgreSQL folds unquoted identifiers. Useful when a source returns
// camelCase names ("createdAt") for a target created without quotes
// (createdat).
func LowercaseColumns() core.Transformer {
	return core.TransformerFunc(func(b core.Batch) (core.Batch, error) {
		cols := make([]core.Column, len(b.Columns))
		for i, c := range b.Columns {
			c.Name = strings.ToLower(c.Name)
			cols[i] = c
		}
		b.Columns = cols
		return b, nil
	})
}

func newLowercaseColumns(Decoder) (core.Transformer, error) { return LowercaseColumns(), nil }

func decode(params Decoder, into any) error {
	if params == nil {
		return nil
	}
	if err := params.Decode(into); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}
	return nil
}
