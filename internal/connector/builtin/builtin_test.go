package builtin

import (
	"strings"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/registry"
)

func TestRegister_ExposesEveryBuiltinDriver(t *testing.T) {
	t.Parallel()
	r := registry.New()
	Register(r, nil)

	for _, d := range []string{Redshift, Postgres, AuroraPostgreSQL, MySQL, AuroraMySQL} {
		if !r.HasSource(d) {
			t.Errorf("missing source driver %q", d)
		}
	}
	for _, d := range []string{Postgres, AuroraPostgreSQL, MySQL, AuroraMySQL, JSONL} {
		if !r.HasSink(d) {
			t.Errorf("missing sink driver %q", d)
		}
	}
	// Writing into Redshift through row-by-row INSERTs would be
	// pathologically slow; it is deliberately not offered as a sink.
	if r.HasSink(Redshift) {
		t.Error("redshift must not be registered as a sink")
	}
}

func TestNormalizePostgresDSN_AcceptsRedshiftScheme(t *testing.T) {
	t.Parallel()
	got, _ := normalizePostgresDSN("redshift://u:p@host:5439/db?sslmode=require")
	if got != "postgres://u:p@host:5439/db?sslmode=require" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeMySQLDSN_ForcesParseTime(t *testing.T) {
	t.Parallel()
	got, err := normalizeMySQLDSN("u:p@tcp(localhost:3306)/db")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "parseTime=true") {
		t.Fatalf("got %q, want parseTime=true", got)
	}
}

func TestFactories_RejectEmptyDSNAndBadOptions(t *testing.T) {
	t.Parallel()
	r := registry.New()
	Register(r, nil)

	if _, err := r.NewSource(registry.Connection{Name: "rs", Driver: Redshift}); err == nil {
		t.Fatal("expected an error for an empty DSN")
	}
	_, err := r.NewSink(registry.Connection{
		Name: "pg", Driver: Postgres, DSN: "postgres://localhost/db",
		Options: map[string]string{"rows_per_statement": "many"},
	})
	if err == nil {
		t.Fatal("expected an error for a non-numeric rows_per_statement")
	}
	if _, err := r.NewSink(registry.Connection{Name: "x", Driver: "oracle"}); err == nil || !strings.Contains(err.Error(), "available") {
		t.Fatalf("unknown driver: err = %v", err)
	}
}
