// Package version holds build metadata, injected at link time:
//
//	go build -ldflags "-X github.com/Jeffersonmf/aurora-exporter/internal/version.Version=v1.0.0 ..."
package version

import "fmt"

// Set by -ldflags at build time (see Makefile).
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String renders every field on one line.
func String() string {
	return fmt.Sprintf("aurora-exporter %s (commit %s, built %s)", Version, Commit, Date)
}
