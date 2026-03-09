package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
)

// MeasureTools contains all measure catalog related tools
type MeasureTools struct {
	clientManager *client.Manager
}

// NewMeasureTools creates a new measure tools instance
func NewMeasureTools(clientManager *client.Manager) *MeasureTools {
	return &MeasureTools{clientManager: clientManager}
}

// List handles measure catalog listing
func (t *MeasureTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.MeasureParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get all available measures from the catalog
	measureGroups, err := auxoClient.ZeroTrust.GetMeasures(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert to response format - flatten all measures from all groups
	allMeasures := FlattenMeasureGroups(*measureGroups)

	// Apply filters if provided
	var filteredMeasures []map[string]interface{}

	// Filter by ID if provided
	if args.ID != "" {
		for _, measure := range allMeasures {
			if measure["name"].(string) == args.ID {
				filteredMeasures = []map[string]interface{}{measure}
				break
			}
		}
	} else {
		filteredMeasures = allMeasures
	}

	// Additional filtering by category/group if provided
	if args.Category != "" {
		var categoryFiltered []map[string]interface{}
		for _, measure := range filteredMeasures {
			if measure["group_name"].(string) == args.Category ||
				measure["group_label"].(string) == args.Category ||
				measure["group_caption"].(string) == args.Category {
				categoryFiltered = append(categoryFiltered, measure)
			}
		}
		filteredMeasures = categoryFiltered
	}

	// Search filter if provided
	if args.SearchQuery != "" {
		var searchFiltered []map[string]interface{}
		searchLower := strings.ToLower(args.SearchQuery)
		for _, measure := range filteredMeasures {
			// Search in name, caption, and explanation (case insensitive)
			if strings.Contains(strings.ToLower(measure["name"].(string)), searchLower) ||
				strings.Contains(strings.ToLower(measure["caption"].(string)), searchLower) ||
				strings.Contains(strings.ToLower(measure["explanation"].(string)), searchLower) {
				searchFiltered = append(searchFiltered, measure)
			}
		}
		filteredMeasures = searchFiltered
	}

	// Convert to JSON
	jsonData, err := json.Marshal(filteredMeasures)
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
