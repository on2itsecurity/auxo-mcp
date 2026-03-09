package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
)

// ContactTools contains all contact related tools
type ContactTools struct {
	clientManager *client.Manager
}

// NewContactTools creates a new contact tools instance
func NewContactTools(clientManager *client.Manager) *ContactTools {
	return &ContactTools{clientManager: clientManager}
}

// List handles contact listing
func (t *ContactTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	allContacts, err := auxoClient.CRM.GetContacts(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allContacts)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}
