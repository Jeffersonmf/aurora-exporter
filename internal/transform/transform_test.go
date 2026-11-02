package transform_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/transform"
)

// jsonParams is a Decoder backed by a JSON literal, standing in for the
// yaml.Node the config package passes in production.
type jsonParams string

func (j jsonParams) Decode(v any) error { return json.Unmarshal([]byte(j), v) }

func batch() core.Batch {
	return core.Batch{
		Columns: []core.Column{{Name: "id"}, {Name: "createdAt"}, {Name: "secret"}},
		Rows:    []core.Row{{1, "2024-01-01", "x"}, {2, "2024-01-02", "y"}},
	}
}

func TestSanitizeUTF8_StripsNULAndInvalidBytes(t *testing.T) {
	t.Parallel()
	b := core.Batch{Columns: []core.Column{{Name: "v"}}, Rows: []core.Row{{"ok"}, {"nul\x00byte"}, {"bad\xffutf8"}, {42}}}

	out, err := transform.SanitizeUTF8().Transform(b)
	if err != nil {
		t.Fatal(err)
	}

	want := []any{"ok", "nulbyte", "badutf8", 42}
	for i, row := range out.Rows {
		if row[0] != want[i] {
			t.Errorf("row %d = %q, want %q", i, row[0], want[i])
		}
	}
}

func TestRenameColumns_RenamesOnlyMappedColumns(t *testing.T) {
	t.Parallel()
	out, err := transform.RenameColumns(map[string]string{"createdAt": "created_at"}).Transform(batch())
	if err != nil {
		t.Fatal(err)
	}
	if got := out.ColumnNames(); !reflect.DeepEqual(got, []string{"id", "created_at", "secret"}) {
		t.Fatalf("columns = %v", got)
	}
}

func TestDropColumns_RemovesColumnAndValues(t *testing.T) {
	t.Parallel()
	out, err := transform.DropColumns("secret").Transform(batch())
	if err != nil {
		t.Fatal(err)
	}
	if got := out.ColumnNames(); !reflect.DeepEqual(got, []string{"id", "createdAt"}) {
		t.Fatalf("columns = %v", got)
	}
	if !reflect.DeepEqual(out.Rows[1], core.Row{2, "2024-01-02"}) {
		t.Fatalf("row = %v", out.Rows[1])
	}
}

func TestDropColumns_RefusesToDropEverything(t *testing.T) {
	t.Parallel()
	if _, err := transform.DropColumns("id", "createdAt", "secret").Transform(batch()); err == nil {
		t.Fatal("expected an error when every column is dropped")
	}
}

func TestLowercaseColumns(t *testing.T) {
	t.Parallel()
	out, _ := transform.LowercaseColumns().Transform(batch())
	if got := out.Columns[1].Name; got != "createdat" {
		t.Fatalf("column = %q, want createdat", got)
	}
}

func TestChain_AppliesInOrder(t *testing.T) {
	t.Parallel()
	chain := transform.Chain(
		transform.RenameColumns(map[string]string{"createdAt": "created_at"}),
		transform.DropColumns("secret"),
	)
	out, err := chain.Transform(batch())
	if err != nil {
		t.Fatal(err)
	}
	if got := out.ColumnNames(); !reflect.DeepEqual(got, []string{"id", "created_at"}) {
		t.Fatalf("columns = %v", got)
	}
}

func TestRegistry_BuildsBuiltinsFromParams(t *testing.T) {
	t.Parallel()
	reg := transform.NewRegistry()

	tr, err := reg.Build("rename_columns", jsonParams(`{"columns":{"id":"pk"}}`))
	if err != nil {
		t.Fatal(err)
	}
	out, _ := tr.Transform(batch())
	if out.Columns[0].Name != "pk" {
		t.Fatalf("column = %q, want pk", out.Columns[0].Name)
	}

	for _, name := range []string{"sanitize_utf8", "lowercase_columns"} {
		if _, err := reg.Build(name, nil); err != nil {
			t.Errorf("Build(%q): %v", name, err)
		}
	}
}

func TestRegistry_RejectsUnknownAndInvalid(t *testing.T) {
	t.Parallel()
	reg := transform.NewRegistry()

	if _, err := reg.Build("nope", nil); err == nil || !strings.Contains(err.Error(), "available") {
		t.Fatalf("unknown transform: err = %v", err)
	}
	if _, err := reg.Build("drop_columns", jsonParams(`{}`)); err == nil {
		t.Fatal("drop_columns without columns must fail")
	}
	if _, err := reg.Build("rename_columns", jsonParams(`{"columns": 3}`)); err == nil {
		t.Fatal("malformed params must fail")
	}
}

func TestRegistry_AcceptsCustomTransformers(t *testing.T) {
	t.Parallel()
	reg := transform.NewRegistry()
	reg.Register("noop", func(transform.Decoder) (core.Transformer, error) {
		return transform.Chain(), nil
	})
	if !reg.Has("noop") {
		t.Fatal("custom transformer not registered")
	}
	if _, err := reg.Build("noop", nil); err != nil {
		t.Fatal(err)
	}
}
