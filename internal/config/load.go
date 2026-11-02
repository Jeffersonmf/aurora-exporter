package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Jeffersonmf/aurora-exporter/internal/core"
)

// Defaults applied when the file leaves a value unset.
const (
	DefaultAddr            = ":8080"
	DefaultShutdownTimeout = 30 * time.Second
	DefaultBatchSize       = 5000
	DefaultAttempts        = 3
	DefaultMode            = core.ModeReplace
)

// LookupEnv resolves ${VAR} references. os.LookupEnv in production,
// a map in tests.
type LookupEnv func(key string) (string, bool)

// Load reads, expands, decodes, defaults and validates the file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // reading the operator-supplied config path is the point
	if err != nil {
		return nil, err
	}
	cfg, err := Parse(data, os.LookupEnv)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

// Parse is Load without the file system.
func Parse(data []byte, env LookupEnv) (*Config, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if err := expandEnv(&root, env); err != nil {
		return nil, err
	}

	// Re-encode the expanded tree and decode it strictly, so a typo
	// like "bacth_size" is an error instead of a silently ignored key.
	expanded, err := yaml.Marshal(&root)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(expanded))
	dec.KnownFields(true)

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}

	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// envRef matches ${NAME} and ${NAME:-default}, preceded optionally by
// a second $ that escapes it ($${NAME} stays literally ${NAME}).
var envRef = regexp.MustCompile(`\$?\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-([^}]*))?\}`)

// expandEnv replaces ${VAR} references in scalar values only, after
// parsing. Expanding the raw text instead would let a secret containing
// YAML syntax (a password with ": " or "#") corrupt the document, and
// would touch keys and comments too. Only the ${...} form is expanded,
// so $1 placeholders and other dollar signs inside SQL are left alone.
func expandEnv(n *yaml.Node, env LookupEnv) error {
	var missing []string
	var walk func(*yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.ScalarNode && strings.Contains(n.Value, "${") {
			n.Value = envRef.ReplaceAllStringFunc(n.Value, func(ref string) string {
				if strings.HasPrefix(ref, "$$") {
					return ref[1:]
				}
				m := envRef.FindStringSubmatch(ref)
				if v, ok := env(m[1]); ok {
					return v
				}
				if strings.Contains(ref, ":-") {
					return m[2]
				}
				missing = append(missing, m[1])
				return ""
			})
			// An expanded value is always a string: a token like
			// "123456" must not turn into an int, nor "yes" into a bool.
			if n.Tag != "" && n.Tag != "!!str" {
				n.Tag = "!!str"
				n.Style = yaml.DoubleQuotedStyle
			}
		}
		for _, c := range n.Content {
			walk(c)
		}
	}
	walk(n)

	if len(missing) > 0 {
		slices.Sort(missing)
		return fmt.Errorf("undefined environment variables: %s (set them, or use ${NAME:-default})",
			strings.Join(slices.Compact(missing), ", "))
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.Server.Addr == "" {
		c.Server.Addr = DefaultAddr
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = Duration(DefaultShutdownTimeout)
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Defaults.BatchSize == 0 {
		c.Defaults.BatchSize = DefaultBatchSize
	}
	if c.Defaults.Attempts == 0 {
		c.Defaults.Attempts = DefaultAttempts
	}

	for i := range c.Pipelines {
		p := &c.Pipelines[i]
		if p.BatchSize == 0 {
			p.BatchSize = c.Defaults.BatchSize
		}
		if p.Attempts == nil {
			a := c.Defaults.Attempts
			p.Attempts = &a
		}
		if p.Timeout == 0 {
			p.Timeout = c.Defaults.Timeout
		}
		if p.Transforms == nil {
			p.Transforms = c.Defaults.Transforms
		}
		if p.Target.Mode == "" {
			p.Target.Mode = string(DefaultMode)
		}
	}
}

var pipelineName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// Validate checks every rule it can without knowing which drivers and
// transformers exist, and reports all violations at once.
func (c *Config) Validate() error {
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if c.Version != CurrentVersion {
		add("version: must be %d, got %d", CurrentVersion, c.Version)
	}
	if !slices.Contains([]string{"debug", "info", "warn", "error"}, c.Logging.Level) {
		add("logging.level: must be debug, info, warn or error, got %q", c.Logging.Level)
	}
	if !slices.Contains([]string{"json", "text"}, c.Logging.Format) {
		add("logging.format: must be json or text, got %q", c.Logging.Format)
	}
	if c.Defaults.BatchSize < 0 {
		add("defaults.batch_size: must be positive")
	}
	if c.Defaults.Attempts < 0 {
		add("defaults.attempts: must be at least 1")
	}

	for name, conn := range c.Connections {
		if conn.Driver == "" {
			add("connections.%s.driver: is required", name)
		}
	}

	if len(c.Pipelines) == 0 {
		add("pipelines: at least one pipeline is required")
	}
	seen := map[string]bool{}
	for i, p := range c.Pipelines {
		at := fmt.Sprintf("pipelines[%d]", i)
		if p.Name != "" {
			at = fmt.Sprintf("pipelines[%d] (%s)", i, p.Name)
		}

		switch {
		case p.Name == "":
			add("%s.name: is required", at)
		case !pipelineName.MatchString(p.Name):
			add("%s.name: must match %s", at, pipelineName)
		case seen[p.Name]:
			add("%s.name: duplicated", at)
		}
		seen[p.Name] = true

		if _, ok := c.Connections[p.Source.Connection]; !ok {
			add("%s.source.connection: unknown connection %q", at, p.Source.Connection)
		}
		if strings.TrimSpace(p.Source.Query) == "" {
			add("%s.source.query: is required", at)
		}
		if _, ok := c.Connections[p.Target.Connection]; !ok {
			add("%s.target.connection: unknown connection %q", at, p.Target.Connection)
		}
		if p.Target.Table == "" {
			add("%s.target.table: is required", at)
		}
		mode := core.WriteMode(p.Target.Mode)
		if err := mode.Validate(); err != nil {
			add("%s.target.mode: %v", at, err)
		}
		if mode == core.ModeUpsert && len(p.Target.Keys) == 0 {
			add("%s.target.keys: upsert requires at least one key column", at)
		}
		if p.BatchSize <= 0 {
			add("%s.batch_size: must be positive", at)
		}
		if p.Attempts != nil && *p.Attempts < 1 {
			add("%s.attempts: must be at least 1", at)
		}
		if every := p.Schedule.Every.Std(); every != 0 && every < time.Second {
			add("%s.schedule.every: must be at least 1s, got %s", at, every)
		}
		for j, t := range p.Transforms {
			if t.Type == "" {
				add("%s.transforms[%d]: type is required", at, j)
			}
		}
	}

	return errors.Join(errs...)
}
