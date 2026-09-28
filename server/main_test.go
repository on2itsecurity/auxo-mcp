package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/config"
)

// listAllTools starts a server with every domain enabled and returns the tools
// a client sees over an in-memory connection. No AUXO API call is made.
func listAllTools(t *testing.T) []*mcp.Tool {
	t.Helper()
	ctx := context.Background()
	cfg := &config.Config{
		APIUrl:          config.DefaultAPIURL,
		ZTToken:         "your-zt-token",
		TicketToken:     "your-ticket-token",
		ServerMode:      "STDIO",
		EnableZeroTrust: true,
		EnableTickets:   true,
	}
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := newServer(cfg).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	var tools []*mcp.Tool
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		tools = append(tools, tool)
	}
	if len(tools) == 0 {
		t.Fatal("no tools registered")
	}
	return tools
}

// TestToolAnnotations enforces what the Claude plugin directory expects: every
// tool has a title and declares readOnlyHint or destructiveHint, and the
// hints match the kind of operation the tool name implies.
func TestToolAnnotations(t *testing.T) {
	alwaysDestructive := map[string]bool{
		"closeCase":      true,
		"escalateCase":   true,
		"deescalateCase": true,
	}

	for _, tool := range listAllTools(t) {
		t.Run(tool.Name, func(t *testing.T) {
			a := tool.Annotations
			if a == nil {
				t.Fatal("tool has no annotations")
			}
			if strings.TrimSpace(tool.Title) == "" || strings.TrimSpace(a.Title) == "" {
				t.Errorf("missing title (tool.title=%q, annotations.title=%q)", tool.Title, a.Title)
			}
			if !a.ReadOnlyHint && a.DestructiveHint == nil {
				t.Error("neither readOnlyHint nor destructiveHint is set")
			}

			name := tool.Name
			switch {
			case strings.HasPrefix(name, "list"), strings.HasPrefix(name, "get"):
				if !a.ReadOnlyHint {
					t.Error("list/get tool must set readOnlyHint: true")
				}
			case strings.HasPrefix(name, "delete"), strings.HasPrefix(name, "remove"), alwaysDestructive[name]:
				if a.ReadOnlyHint || a.DestructiveHint == nil || !*a.DestructiveHint {
					t.Error("tool must set readOnlyHint: false and destructiveHint: true")
				}
			case strings.HasPrefix(name, "create"), strings.HasPrefix(name, "update"), strings.HasPrefix(name, "add"):
				if a.ReadOnlyHint || a.DestructiveHint == nil {
					t.Error("create/update tool must set readOnlyHint: false and an explicit destructiveHint")
				}
			}
		})
	}
}
