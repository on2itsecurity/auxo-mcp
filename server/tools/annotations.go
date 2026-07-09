package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// ReadOnlyAnnotations returns annotations for a tool that only reads data
// from the AUXO API and never modifies it.
func ReadOnlyAnnotations(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		Title:         title,
		ReadOnlyHint:  true,
		OpenWorldHint: boolPtr(false),
	}
}

// WriteAnnotations returns annotations for a tool that modifies data in the
// AUXO API. destructive indicates the tool can delete or overwrite existing
// data (as opposed to purely additive changes); idempotent indicates that
// repeating the call with the same arguments has no additional effect.
func WriteAnnotations(title string, destructive, idempotent bool) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		Title:           title,
		DestructiveHint: boolPtr(destructive),
		IdempotentHint:  idempotent,
		OpenWorldHint:   boolPtr(false),
	}
}

func boolPtr(b bool) *bool { return &b }
