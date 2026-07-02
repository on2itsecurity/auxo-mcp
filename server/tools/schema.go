package tools

import (
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InputSchemaFor builds a JSON Schema for the given tool input type and removes
// the "null" member from any "type" arrays that the reflection-based generator
// emits for pointer (*int, *bool, *float64, ...) and slice ([]string) fields.
//
// The go-sdk / jsonschema-go reflector renders an optional pointer or slice
// field as:
//
//	"confidentiality": { "type": ["null", "integer"] }
//
// The "type": [...] array form is valid JSON Schema (2020-12) but is not
// universally handled by MCP client tool-calling layers: some clients fail to
// recognize it as a typed field and fall back to serializing every argument as
// a string ("85" instead of 85, "true" instead of true). The AUXO server then
// correctly rejects the mistyped string, so every numeric/boolean/array field
// fails while plain string fields keep working.
//
// Optional fields are already modeled by omission from "required", so an
// explicit "null" carries no extra information for our tools. Dropping the
// "null" member collapses the schema back to a single "type": X, which every
// client understands, without touching field names, descriptions, or the
// server-side validation/request handling (the Go structs keep their pointer
// types so nil still means "not set").
func InputSchemaFor[In any]() (*jsonschema.Schema, error) {
	schema, err := jsonschema.For[In](nil)
	if err != nil {
		return nil, err
	}
	stripNullType(schema)
	return schema, nil
}

// stripNullType recursively rewrites schema so that no "type" array contains
// "null". When removing "null" leaves exactly one type, the schema is collapsed
// to the scalar Type form ("type": "integer") instead of a single-element
// array. Nested schemas (properties, items, combinators, $defs) are handled too.
func stripNullType(s *jsonschema.Schema) {
	if s == nil {
		return
	}

	if len(s.Types) > 0 {
		filtered := make([]string, 0, len(s.Types))
		for _, t := range s.Types {
			if t != "null" {
				filtered = append(filtered, t)
			}
		}
		switch len(filtered) {
		case 1:
			// Collapse ["null","integer"] -> "integer".
			s.Type = filtered[0]
			s.Types = nil
		case 0:
			// Only "null" was present; nothing sensible to collapse to. Leave
			// the field untouched rather than producing an invalid schema.
		default:
			s.Types = filtered
		}
	}

	stripNullType(s.Items)
	stripNullType(s.AdditionalProperties)
	stripNullType(s.AdditionalItems)
	for _, p := range s.Properties {
		stripNullType(p)
	}
	for _, p := range s.PatternProperties {
		stripNullType(p)
	}
	for _, p := range s.PrefixItems {
		stripNullType(p)
	}
	for _, a := range s.AllOf {
		stripNullType(a)
	}
	for _, a := range s.AnyOf {
		stripNullType(a)
	}
	for _, a := range s.OneOf {
		stripNullType(a)
	}
	for _, d := range s.Defs {
		stripNullType(d)
	}
}

// AddTool registers a typed tool on server, exactly like mcp.AddTool, except
// that the reflected input schema has its nullable "type" arrays collapsed to
// single types (see InputSchemaFor). If the caller already supplied an explicit
// InputSchema on t, it is respected as-is.
func AddTool[In, Out any](server *mcp.Server, t *mcp.Tool, h mcp.ToolHandlerFor[In, Out]) {
	if t.InputSchema == nil {
		schema, err := InputSchemaFor[In]()
		if err != nil {
			panic(fmt.Sprintf("AddTool: building input schema for %q: %v", t.Name, err))
		}
		t.InputSchema = schema
	}
	mcp.AddTool(server, t, h)
}
