// Package transform holds the built-in core.Transformer
// implementations and a Registry that builds them by name from
// configuration.
package transform

import (
	"fmt"
	"slices"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// Decoder decodes a transformer's parameters into a struct. yaml.Node
// satisfies it, but this package does not depend on any configuration
// format: the caller decides where parameters come from.
type Decoder interface {
	Decode(v any) error
}

// Factory builds a Transformer from its parameters.
type Factory func(params Decoder) (core.Transformer, error)

// Registry maps transformer type names to factories. New transformers
// are added by registering a factory, never by editing a switch
// statement somewhere else (Open/Closed).
type Registry struct {
	factories map[string]Factory
}

// NewRegistry returns a Registry with every built-in transformer.
func NewRegistry() *Registry {
	r := &Registry{factories: map[string]Factory{}}
	r.Register("sanitize_utf8", newSanitizeUTF8)
	r.Register("rename_columns", newRenameColumns)
	r.Register("drop_columns", newDropColumns)
	r.Register("lowercase_columns", newLowercaseColumns)
	return r
}

// Register adds or replaces the factory for name.
func (r *Registry) Register(name string, f Factory) {
	r.factories[name] = f
}

// Build creates the transformer registered as name.
func (r *Registry) Build(name string, params Decoder) (core.Transformer, error) {
	f, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("unknown transform %q (available: %v)", name, r.Names())
	}
	t, err := f(params)
	if err != nil {
		return nil, fmt.Errorf("transform %q: %w", name, err)
	}
	return t, nil
}

// Has reports whether name is registered.
func (r *Registry) Has(name string) bool {
	_, ok := r.factories[name]
	return ok
}

// Names lists registered transformer names, sorted.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.factories))
	for n := range r.factories {
		names = append(names, n)
	}
	slices.Sort(names)
	return names
}

// Chain composes transformers left to right. An empty chain returns
// batches unchanged.
func Chain(ts ...core.Transformer) core.Transformer {
	return core.TransformerFunc(func(b core.Batch) (core.Batch, error) {
		var err error
		for _, t := range ts {
			if b, err = t.Transform(b); err != nil {
				return core.Batch{}, err
			}
		}
		return b, nil
	})
}
