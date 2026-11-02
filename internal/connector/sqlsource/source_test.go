package sqlsource_test

import (
	"testing"

	"github.com/Jeffersonmf/aurora-exporter/internal/connector/sqlsource"
)

func TestBinaryTypes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		fn   sqlsource.BinaryTypes
		typ  string
		want bool
	}{
		{sqlsource.PostgresBinaryTypes, "BYTEA", true},
		{sqlsource.PostgresBinaryTypes, "varbyte", true},
		{sqlsource.PostgresBinaryTypes, "NUMERIC", false},
		{sqlsource.PostgresBinaryTypes, "JSONB", false},
		{sqlsource.MySQLBinaryTypes, "BLOB", true},
		{sqlsource.MySQLBinaryTypes, "VARBINARY", true},
		{sqlsource.MySQLBinaryTypes, "DECIMAL", false},
		{sqlsource.MySQLBinaryTypes, "VARCHAR", false},
	}
	for _, tc := range cases {
		if got := tc.fn(tc.typ); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.typ, got, tc.want)
		}
	}
}

// Value normalization against a real driver (NUMERIC arriving as
// []byte, BYTEA staying binary, timestamps as time.Time) is covered by
// the integration suite: test/integration.
