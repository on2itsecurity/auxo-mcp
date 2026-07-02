package tools

import (
	"encoding/json"
	"testing"

	"github.com/on2itsecurity/auxo-mcp/server/types"
)

// TestPrintGeneratedSchemas is a debug harness (not an assertion) that prints
// the generated input schemas so a maintainer can eyeball that the affected
// fields now use a single "type" string instead of a "type": ["null", X] array.
//
// Run it with:
//
//	go test ./server/tools/ -run TestPrintGeneratedSchemas -v
func TestPrintGeneratedSchemas(t *testing.T) {
	for _, tc := range []struct {
		tool   string
		schema func() (any, error)
	}{
		{"createProtectSurface", func() (any, error) { return InputSchemaFor[types.ProtectSurfaceParams]() }},
		{"createState", func() (any, error) { return InputSchemaFor[types.StateParams]() }},
	} {
		s, err := tc.schema()
		if err != nil {
			t.Fatalf("%s: %v", tc.tool, err)
		}
		out, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			t.Fatalf("%s: marshal: %v", tc.tool, err)
		}
		t.Logf("\n=== %s inputSchema ===\n%s", tc.tool, out)
	}
}
