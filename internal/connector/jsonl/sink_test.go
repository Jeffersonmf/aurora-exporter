package jsonl_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/connector/jsonl"
	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

var (
	cols  = []core.Column{{Name: "id"}, {Name: "name"}}
	batch = core.Batch{Columns: cols, Rows: []core.Row{{1, "a"}, {2, nil}}}
	ctx   = context.Background()
)

func TestSink_StdoutWritesOneObjectPerRow(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	w, err := jsonl.New(jsonl.Stdout, &out).Open(ctx, core.Target{Table: "t", Mode: core.ModeReplace}, cols)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(ctx, batch); err != nil {
		t.Fatal(err)
	}
	if err := w.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	want := `{"id":1,"name":"a"}` + "\n" + `{"id":2,"name":null}` + "\n"
	if out.String() != want {
		t.Fatalf("output =\n%s\nwant\n%s", out.String(), want)
	}
}

func TestSink_FileReplaceIsAtomic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "{table}.jsonl")
	final := filepath.Join(dir, "brands.jsonl")
	if err := os.WriteFile(final, []byte("previous\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sink := jsonl.New(path, nil)
	target := core.Target{Table: "brands", Mode: core.ModeReplace}

	// A rolled-back run leaves the previous file and no temp files.
	w, _ := sink.Open(ctx, target, cols)
	_, _ = w.Write(ctx, batch)
	if err := w.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	assertFile(t, final, "previous\n")
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("temp file left behind: %v", entries)
	}

	// A committed run replaces it.
	w, _ = sink.Open(ctx, target, cols)
	_, _ = w.Write(ctx, batch)
	if err := w.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(final)
	if strings.Count(string(data), "\n") != 2 || strings.Contains(string(data), "previous") {
		t.Fatalf("file = %q", data)
	}
}

func TestSink_FileAppend(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "out.jsonl")
	sink := jsonl.New(path, nil)

	for range 2 {
		w, err := sink.Open(ctx, core.Target{Table: "t", Mode: core.ModeAppend}, cols)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write(ctx, batch)
		if err := w.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := os.ReadFile(path)
	if got := strings.Count(string(data), "\n"); got != 4 {
		t.Fatalf("lines = %d, want 4", got)
	}
}

func TestSink_RejectsUpsert(t *testing.T) {
	t.Parallel()
	if _, err := jsonl.New(jsonl.Stdout, nil).Open(ctx, core.Target{Table: "t", Mode: core.ModeUpsert}, cols); err == nil {
		t.Fatal("expected upsert to be rejected")
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}
