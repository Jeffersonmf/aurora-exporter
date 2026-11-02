// Package config loads and validates the aurora-exporter YAML file.
//
// It only describes and checks the file's structure. Turning it into
// running pipelines (resolving drivers, building connectors and
// transformers) is the app package's job, so this package never needs
// to change when a new driver or transformer is added.
package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// CurrentVersion is the only config file version this build reads.
const CurrentVersion = 1

// Config is the root of the configuration file.
type Config struct {
	Version     int                   `yaml:"version"`
	Server      Server                `yaml:"server"`
	Logging     Logging               `yaml:"logging"`
	Defaults    Defaults              `yaml:"defaults"`
	Connections map[string]Connection `yaml:"connections"`
	Pipelines   []Pipeline            `yaml:"pipelines"`
}

// Server configures the HTTP API used by the serve command.
type Server struct {
	Addr string `yaml:"addr"`
	// AuthToken, when set, is required as a Bearer token on every
	// /v1 endpoint. Health probes stay open.
	AuthToken       string   `yaml:"auth_token"`
	ShutdownTimeout Duration `yaml:"shutdown_timeout"`
}

// Logging configures the process logger.
type Logging struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

// Defaults apply to every pipeline that does not override them.
type Defaults struct {
	BatchSize  int         `yaml:"batch_size"`
	Attempts   int         `yaml:"attempts"`
	Timeout    Duration    `yaml:"timeout"`
	Transforms []Transform `yaml:"transforms"`
}

// Connection is a named database (or other endpoint) that pipelines
// read from or write to.
type Connection struct {
	Driver          string            `yaml:"driver"`
	DSN             string            `yaml:"dsn"`
	MaxOpenConns    int               `yaml:"max_open_conns"`
	MaxIdleConns    int               `yaml:"max_idle_conns"`
	ConnMaxLifetime Duration          `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime Duration          `yaml:"conn_max_idle_time"`
	Options         map[string]string `yaml:"options"`
}

// Pipeline is one export from a source query to a target table.
type Pipeline struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	// Enabled defaults to true. A disabled pipeline is still listed
	// and can be triggered manually, it is just never scheduled.
	Enabled   *bool    `yaml:"enabled"`
	Source    Source   `yaml:"source"`
	Target    Target   `yaml:"target"`
	Schedule  Schedule `yaml:"schedule"`
	BatchSize int      `yaml:"batch_size"`
	// Attempts is how many times a scheduled run is tried before it is
	// reported as failed (1 = no retry).
	Attempts *int     `yaml:"attempts"`
	Timeout  Duration `yaml:"timeout"`
	// Transforms replaces defaults.transforms when present, including
	// when present and empty (transforms: []).
	Transforms []Transform `yaml:"transforms"`
}

// IsEnabled reports whether the pipeline should be scheduled.
func (p Pipeline) IsEnabled() bool { return p.Enabled == nil || *p.Enabled }

// Source is where a pipeline reads.
type Source struct {
	Connection string `yaml:"connection"`
	Query      string `yaml:"query"`
}

// Target is where a pipeline writes.
type Target struct {
	Connection string   `yaml:"connection"`
	Table      string   `yaml:"table"`
	Mode       string   `yaml:"mode"`
	Keys       []string `yaml:"keys"`
}

// Schedule makes a pipeline run periodically under the serve command.
// The zero value means "manual only".
type Schedule struct {
	Every Duration `yaml:"every"`
}

// Transform is one entry of a transforms list. It accepts both a bare
// name (- sanitize_utf8) and a mapping with a type and parameters
// (- type: rename_columns / columns: {...}).
type Transform struct {
	Type string
	// Params holds the entry's raw YAML so each transformer decodes its
	// own parameters; this package does not need to know them.
	Params *yaml.Node
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (t *Transform) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		t.Type = n.Value
		return nil
	case yaml.MappingNode:
		var head struct {
			Type string `yaml:"type"`
		}
		if err := n.Decode(&head); err != nil {
			return err
		}
		t.Type, t.Params = head.Type, n
		return nil
	default:
		return fmt.Errorf("line %d: a transform is either a name or a mapping with a type", n.Line)
	}
}

// MarshalYAML implements yaml.Marshaler, so a Config can be printed
// back (the validate command does).
func (t Transform) MarshalYAML() (any, error) {
	if t.Params != nil {
		return t.Params, nil
	}
	return t.Type, nil
}

// Duration is a time.Duration written as a Go duration string ("90s",
// "15m", "1h30m") in YAML.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(n *yaml.Node) error {
	if n.Value == "" || n.Value == "0" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(n.Value)
	if err != nil {
		return fmt.Errorf("line %d: invalid duration %q (use values like 30s, 15m, 1h)", n.Line, n.Value)
	}
	*d = Duration(parsed)
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (d Duration) MarshalYAML() (any, error) { return time.Duration(d).String(), nil }

// Std returns d as a time.Duration.
func (d Duration) Std() time.Duration { return time.Duration(d) }
