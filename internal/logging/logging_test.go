package logging_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/logging"
)

func TestNew(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	log, err := logging.New(&buf, "warn", "json")
	if err != nil {
		t.Fatal(err)
	}
	log.Info("hidden")
	log.Warn("shown", "k", "v")
	if out := buf.String(); strings.Contains(out, "hidden") || !strings.Contains(out, `"k":"v"`) {
		t.Fatalf("output = %s", out)
	}

	if _, err := logging.New(&buf, "loud", "json"); err == nil {
		t.Fatal("expected invalid level error")
	}
	if _, err := logging.New(&buf, "info", "xml"); err == nil {
		t.Fatal("expected invalid format error")
	}
}
