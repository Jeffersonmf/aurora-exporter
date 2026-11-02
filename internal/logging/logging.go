// Package logging builds the process-wide structured logger.
package logging

import (
	"fmt"
	"io"
	"log/slog"
)

// New returns a slog.Logger writing to w at level ("debug", "info",
// "warn", "error") in format ("json" or "text").
func New(w io.Writer, level, format string) (*slog.Logger, error) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("logging level: %w", err)
	}
	opts := &slog.HandlerOptions{Level: lvl}

	switch format {
	case "json", "":
		return slog.New(slog.NewJSONHandler(w, opts)), nil
	case "text":
		return slog.New(slog.NewTextHandler(w, opts)), nil
	default:
		return nil, fmt.Errorf("logging format: must be json or text, got %q", format)
	}
}
