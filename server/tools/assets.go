package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
)

// AssetTools contains all asset related tools
type AssetTools struct {
	clientManager *client.Manager
}

// NewAssetTools creates a new asset tools instance
func NewAssetTools(clientManager *client.Manager) *AssetTools {
	return &AssetTools{clientManager: clientManager}
}

// List handles asset listing
func (t *AssetTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	allAssets, err := auxoClient.Asset.GetAssets(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allAssets)
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
