// Package jsonl implements core.Sink as newline-delimited JSON, one
// object per row. It is useful for dry runs, debugging a query's output
// and feeding tools that read JSON lines, and it exists to show that a
// Sink does not have to be a database.
package jsonl

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// Stdout is the path that selects standard output instead of a file.
const Stdout = "-"

// TablePlaceholder in a path is replaced by the target table, so one
// connection can serve several pipelines without them overwriting
// each other's files.
const TablePlaceholder = "{table}"

// Sink writes JSON lines to a file or to stdout.
//
// File output is atomic: rows go to a temporary file in the same
// directory, renamed over the final path only on Commit, so readers
// never see a half-written file and a failed run leaves the previous
// file intact. Stdout output streams as it goes and cannot be rolled
// back.
type Sink struct {
	path   string
	stdout io.Writer
}

// New returns a Sink writing to path ("-" for stdout).
func New(path string, stdout io.Writer) *Sink {
	return &Sink{path: path, stdout: stdout}
}

// Open implements core.Sink. Write modes other than append and replace
// are meaningless for a file and are rejected.
func (s *Sink) Open(_ context.Context, target core.Target, columns []core.Column) (core.Writer, error) {
	if target.Mode != core.ModeAppend && target.Mode != core.ModeReplace {
		return nil, fmt.Errorf("jsonl sink supports modes %q and %q, not %q", core.ModeAppend, core.ModeReplace, target.Mode)
	}

	names := make([]string, len(columns))
	for i, c := range columns {
		names[i] = c.Name
	}

	if s.path == Stdout {
		return &writer{out: bufio.NewWriter(s.stdout), columns: names}, nil
	}

	final := strings.ReplaceAll(s.path, TablePlaceholder, target.Table)
	if target.Mode == core.ModeAppend {
		f, err := os.OpenFile(final, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) //nolint:gosec // path comes from the operator's config
		if err != nil {
			return nil, err
		}
		return &writer{out: bufio.NewWriter(f), file: f, columns: names}, nil
	}

	tmp, err := os.CreateTemp(filepath.Dir(final), filepath.Base(final)+".*.tmp")
	if err != nil {
		return nil, err
	}
	return &writer{out: bufio.NewWriter(tmp), file: tmp, final: final, columns: names}, nil
}

type writer struct {
	out     *bufio.Writer
	file    *os.File
	final   string // non-empty: rename file to final on commit
	columns []string
	done    bool
}

func (w *writer) Write(_ context.Context, b core.Batch) (int64, error) {
	enc := json.NewEncoder(w.out)
	var n int64
	for _, row := range b.Rows {
		obj := make(map[string]any, len(w.columns))
		for i, name := range w.columns {
			if i < len(row) {
				obj[name] = row[i]
			}
		}
		if err := enc.Encode(obj); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (w *writer) Commit(context.Context) error {
	if w.done {
		return errors.New("writer already finished")
	}
	w.done = true

	if err := w.out.Flush(); err != nil {
		return errors.Join(err, w.discard())
	}
	if w.file == nil {
		return nil
	}
	if err := w.file.Sync(); err != nil {
		return errors.Join(err, w.discard())
	}
	if err := w.file.Close(); err != nil {
		return errors.Join(err, w.discard())
	}
	if w.final != "" {
		return os.Rename(w.file.Name(), w.final)
	}
	return nil
}

func (w *writer) Rollback(context.Context) error {
	if w.done {
		return nil
	}
	w.done = true
	return w.discard()
}

// discard closes the file and, for atomic writes, removes the
// temporary file. Appended rows already on disk stay there.
func (w *writer) discard() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	if errors.Is(err, os.ErrClosed) {
		err = nil
	}
	if w.final != "" {
		if rerr := os.Remove(w.file.Name()); rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
			err = errors.Join(err, rerr)
		}
	}
	return err
}
