package registry_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
	"github.com/Jeffersonmf/aurora-exporter/internal/core/coretest"
	"github.com/Jeffersonmf/aurora-exporter/internal/registry"
)

func TestRegistry_RoleSpecificLookup(t *testing.T) {
	t.Parallel()
	r := registry.New()
	r.RegisterSource("b-src", func(registry.Connection) (core.Source, error) { return &coretest.Source{}, nil })
	r.RegisterSource("a-src", func(registry.Connection) (core.Source, error) { return &coretest.Source{}, nil })
	r.RegisterSink("sink", func(registry.Connection) (core.Sink, error) { return &coretest.Sink{}, nil })

	if got := r.SourceDrivers(); !reflect.DeepEqual(got, []string{"a-src", "b-src"}) {
		t.Fatalf("SourceDrivers = %v, want sorted", got)
	}
	if !r.HasSink("sink") || r.HasSink("a-src") || r.HasSource("sink") {
		t.Fatal("sources and sinks must be looked up independently")
	}
	if _, err := r.NewSource(registry.Connection{Name: "c", Driver: "a-src"}); err != nil {
		t.Fatal(err)
	}
	_, err := r.NewSink(registry.Connection{Name: "c", Driver: "a-src"})
	if err == nil || !strings.Contains(err.Error(), "cannot be used as a target") {
		t.Fatalf("err = %v", err)
	}
}

func TestConnection_IntOption(t *testing.T) {
	t.Parallel()
	c := registry.Connection{Name: "c", Options: map[string]string{"n": "42", "bad": "x"}}
	if v, err := c.IntOption("n", 1); err != nil || v != 42 {
		t.Fatalf("n = %d, %v", v, err)
	}
	if v, _ := c.IntOption("missing", 7); v != 7 {
		t.Fatalf("fallback = %d", v)
	}
	if _, err := c.IntOption("bad", 0); err == nil {
		t.Fatal("expected parse error")
	}
}
