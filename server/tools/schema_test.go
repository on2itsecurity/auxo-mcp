package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/on2itsecurity/auxo-mcp/server/types"
)

// resolveInput builds the cleaned input schema for In and resolves it the same
// way the go-sdk does before validating tool-call arguments.
func resolveInput[In any](t *testing.T) *jsonschema.Resolved {
	t.Helper()
	schema, err := InputSchemaFor[In]()
	if err != nil {
		t.Fatalf("InputSchemaFor: %v", err)
	}
	resolved, err := schema.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	return resolved
}

// validate mirrors the SDK's applySchema: apply defaults, then validate a
// decoded JSON object against the resolved schema.
func validate(t *testing.T, resolved *jsonschema.Resolved, argsJSON string) error {
	t.Helper()
	v := make(map[string]any)
	if err := json.Unmarshal([]byte(argsJSON), &v); err != nil {
		t.Fatalf("unmarshal args: %v", err)
	}
	if err := resolved.ApplyDefaults(&v); err != nil {
		return err
	}
	return resolved.Validate(&v)
}

// prop fetches a named property schema, failing the test if it is missing.
func prop(t *testing.T, s *jsonschema.Schema, name string) *jsonschema.Schema {
	t.Helper()
	p, ok := s.Properties[name]
	if !ok {
		t.Fatalf("property %q not found in schema", name)
	}
	return p
}

// TestInputSchemaHasNoNullTypeArrays is the core regression guard: the
// generated schema for nullable numeric/boolean/array fields must NOT use the
// "type": ["null", X] array form, which some MCP clients mishandle by
// stringifying every argument. See InputSchemaFor.
func TestInputSchemaHasNoNullTypeArrays(t *testing.T) {
	cases := []struct {
		name   string
		schema func() (*jsonschema.Schema, error)
		// field -> expected single "type"
		want map[string]string
	}{
		{
			name:   "ProtectSurfaceParams",
			schema: InputSchemaFor[types.ProtectSurfaceParams],
			want: map[string]string{
				"confidentiality":     "integer",
				"integrity":           "integer",
				"availability":        "integer",
				"relevance":           "integer",
				"in_control_boundary": "boolean",
				"in_zero_trust_focus": "boolean",
				"data_tags":           "array",
				"compliance_tags":     "array",
				"ids":                 "array",
			},
		},
		{
			name:   "StateParams",
			schema: InputSchemaFor[types.StateParams],
			want: map[string]string{
				"content":             "array",
				"exists_on_asset_ids": "array",
				"ids":                 "array",
			},
		},
		{
			name:   "LocationParams",
			schema: InputSchemaFor[types.LocationParams],
			want: map[string]string{
				"latitude":  "number",
				"longitude": "number",
				"ids":       "array",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := tc.schema()
			if err != nil {
				t.Fatalf("build schema: %v", err)
			}

			// No property (at any depth) may keep a "type" array containing null.
			raw, err := json.Marshal(s)
			if err != nil {
				t.Fatalf("marshal schema: %v", err)
			}
			if strings.Contains(string(raw), `"null"`) {
				t.Errorf("schema still contains a \"null\" type member:\n%s", raw)
			}

			for field, wantType := range tc.want {
				p := prop(t, s, field)
				if len(p.Types) > 0 {
					t.Errorf("field %q uses type array %v, want single type %q", field, p.Types, wantType)
				}
				if p.Type != wantType {
					t.Errorf("field %q has type %q, want %q", field, p.Type, wantType)
				}
			}
		})
	}
}

// TestTypedArgumentsAccepted asserts that properly typed int/bool/array
// arguments validate successfully against the resolved input schema — the
// behavior that broke when clients stringified arguments to satisfy the
// "type": ["null", X] form.
func TestTypedArgumentsAccepted(t *testing.T) {
	t.Run("createProtectSurface typed ints/bools/arrays", func(t *testing.T) {
		resolved := resolveInput[types.ProtectSurfaceParams](t)
		args := `{
			"name": "Crown Jewels",
			"confidentiality": 85,
			"integrity": 5,
			"availability": 3,
			"relevance": 100,
			"in_control_boundary": true,
			"in_zero_trust_focus": false,
			"data_tags": ["PII", "PCI"],
			"compliance_tags": ["GDPR"]
		}`
		if err := validate(t, resolved, args); err != nil {
			t.Fatalf("typed arguments were rejected: %v", err)
		}
	})

	t.Run("createState typed array content", func(t *testing.T) {
		resolved := resolveInput[types.StateParams](t)
		args := `{
			"protectsurface_id": "ps-123",
			"location_id": "loc-1",
			"content_type": "ipv4",
			"content": ["10.10.0.1", "10.10.0.2"],
			"exists_on_asset_ids": ["asset-1"]
		}`
		if err := validate(t, resolved, args); err != nil {
			t.Fatalf("typed arguments were rejected: %v", err)
		}
	})
}

// TestStringifiedArgumentsRejected proves the schema is genuinely typed: the
// stringified forms that a broken client would send ("85", "true",
// "[\"10.10.0.1\"]") must fail validation, so we know the single-type schema is
// real and not silently accepting everything.
func TestStringifiedArgumentsRejected(t *testing.T) {
	resolved := resolveInput[types.ProtectSurfaceParams](t)

	stringified := map[string]string{
		"stringified int":   `{"name":"x","confidentiality":"85"}`,
		"stringified bool":  `{"name":"x","in_control_boundary":"true"}`,
		"stringified array": `{"name":"x","data_tags":"[\"PII\"]"}`,
	}
	for name, args := range stringified {
		t.Run(name, func(t *testing.T) {
			if err := validate(t, resolved, args); err == nil {
				t.Errorf("expected stringified argument to be rejected, but it passed: %s", args)
			}
		})
	}
}

// TestOmittedOptionalFieldsAccepted confirms optional fields may be omitted
// entirely (the intended way to express "not set" now that null is gone).
func TestOmittedOptionalFieldsAccepted(t *testing.T) {
	resolved := resolveInput[types.ProtectSurfaceParams](t)
	if err := validate(t, resolved, `{"name":"only-name"}`); err != nil {
		t.Fatalf("minimal arguments were rejected: %v", err)
	}
}
