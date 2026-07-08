// Package apps embeds the MCP Apps (interactive UI) resources served by the
// AUXO MCP server. MCP Apps is the official MCP extension
// "io.modelcontextprotocol/ui" (spec 2026-01-26): a tool can reference a
// ui:// HTML resource via _meta, and hosts that support the extension render
// that HTML in a sandboxed iframe which talks JSON-RPC to the host over
// postMessage. Hosts without the extension ignore the _meta and fall back to
// the tool's regular text content.
package apps

import (
	"context"
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// UIExtensionID is the MCP extension identifier for MCP Apps.
const UIExtensionID = "io.modelcontextprotocol/ui"

// AppMIMEType is the MIME type mandated by the MCP Apps extension for UI resources.
const AppMIMEType = "text/html;profile=mcp-app"

// ReadinessAppURI is the ui:// resource URI of the readiness assessment app.
const ReadinessAppURI = "ui://auxo/readiness-assessment.html"

//go:embed readiness.html
var readinessHTML string

// ToolMeta returns the tool _meta that links a tool to a ui:// resource. Both
// the current nested key (_meta.ui.resourceUri) and the deprecated flat key
// (_meta["ui/resourceUri"]) are set, matching the reference SDK, because some
// hosts (early Claude releases) still read the flat form.
func ToolMeta(resourceURI string) mcp.Meta {
	return mcp.Meta{
		"ui":             map[string]any{"resourceUri": resourceURI},
		"ui/resourceUri": resourceURI,
	}
}

// RegisterReadinessApp registers the readiness assessment app as a resource.
// The app is fully self-contained (inline CSS/JS, data: images) so no CSP
// declaration is needed; the extension's restrictive default applies.
func RegisterReadinessApp(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		URI:         ReadinessAppURI,
		Name:        "readiness-assessment-app",
		Title:       "Zero Trust Readiness Assessment",
		Description: "Interactive Zero Trust readiness assessment questionnaire (MCP App)",
		MIMEType:    AppMIMEType,
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      ReadinessAppURI,
				MIMEType: AppMIMEType,
				Text:     readinessHTML,
				Meta: mcp.Meta{
					"ui": map[string]any{"prefersBorder": true},
				},
			}},
		}, nil
	})
}
