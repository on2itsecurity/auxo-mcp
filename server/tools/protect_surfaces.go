package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
	"github.com/on2itsecurity/go-auxo/v2/zerotrust"
)

// ProtectSurfaceTools contains all protect surface related tools
type ProtectSurfaceTools struct {
	clientManager *client.Manager
}

// NewProtectSurfaceTools creates a new protect surface tools instance
func NewProtectSurfaceTools(clientManager *client.Manager) *ProtectSurfaceTools {
	return &ProtectSurfaceTools{clientManager: clientManager}
}

// Create handles protect surface creation
func (t *ProtectSurfaceTools) Create(ctx context.Context, req *mcp.CallToolRequest, args types.ProtectSurfaceParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for creation
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required for creating a protect surface")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Build the protect surface object with all provided fields
	protectSurface := zerotrust.ProtectSurface{
		Name:                    args.Name,
		Description:             args.Description,
		UniquenessKey:           args.UniquenessKey,
		MainContactPersonID:     args.MainContactPersonID,
		SecurityContactPersonID: args.SecurityContactPersonID,
		DataTags:                args.DataTags,
		ComplianceTags:          args.ComplianceTags,
		CustomerLabels:          args.CustomerLabels,
	}

	// Handle pointer fields for bool and int values
	if args.InControlBoundary != nil {
		protectSurface.InControlBoundary = *args.InControlBoundary
	}
	if args.InZeroTrustFocus != nil {
		protectSurface.InZeroTrustFocus = *args.InZeroTrustFocus
	}
	if args.Relevance != nil {
		protectSurface.Relevance = *args.Relevance
	}
	if args.Confidentiality != nil {
		protectSurface.Confidentiality = *args.Confidentiality
	}
	if args.Integrity != nil {
		protectSurface.Integrity = *args.Integrity
	}
	if args.Availability != nil {
		protectSurface.Availability = *args.Availability
	}

	ps, err := auxoClient.ZeroTrust.CreateProtectSurfaceByObject(ctx, protectSurface, false)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert the created protect surface to JSON
	jsonData, err := json.Marshal(ps)
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

// protectSurfaceSummary is a lightweight representation for list responses
type protectSurfaceSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Relevance int    `json:"relevance"`
}

// List handles protect surface listing
func (t *ProtectSurfaceTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	allProtectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Return lightweight summaries to reduce context window usage
	summaries := make([]protectSurfaceSummary, len(allProtectSurfaces))
	for i, ps := range allProtectSurfaces {
		summaries[i] = protectSurfaceSummary{
			ID:        ps.ID,
			Name:      ps.Name,
			Relevance: ps.Relevance,
		}
	}

	jsonData, err := json.Marshal(summaries)
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

// Update handles protect surface updates
func (t *ProtectSurfaceTools) Update(ctx context.Context, req *mcp.CallToolRequest, args types.ProtectSurfaceParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for update
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for updating a protect surface")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the current protect surface to merge with updates
	currentPS, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Update fields if provided, otherwise keep existing values
	updatePS := *currentPS

	// Update string fields if provided
	if args.Name != "" {
		updatePS.Name = args.Name
	}
	if args.Description != "" {
		updatePS.Description = args.Description
	}
	if args.UniquenessKey != "" {
		updatePS.UniquenessKey = args.UniquenessKey
	}
	if args.MainContactPersonID != "" {
		updatePS.MainContactPersonID = args.MainContactPersonID
	}
	if args.SecurityContactPersonID != "" {
		updatePS.SecurityContactPersonID = args.SecurityContactPersonID
	}

	// Update pointer fields for bool and int values
	if args.InControlBoundary != nil {
		updatePS.InControlBoundary = *args.InControlBoundary
	}
	if args.InZeroTrustFocus != nil {
		updatePS.InZeroTrustFocus = *args.InZeroTrustFocus
	}
	if args.Relevance != nil {
		updatePS.Relevance = *args.Relevance
	}
	if args.Confidentiality != nil {
		updatePS.Confidentiality = *args.Confidentiality
	}
	if args.Integrity != nil {
		updatePS.Integrity = *args.Integrity
	}
	if args.Availability != nil {
		updatePS.Availability = *args.Availability
	}

	// Update slice fields if provided
	if args.DataTags != nil {
		updatePS.DataTags = args.DataTags
	}
	if args.ComplianceTags != nil {
		updatePS.ComplianceTags = args.ComplianceTags
	}

	// Update map fields if provided
	if args.CustomerLabels != nil {
		updatePS.CustomerLabels = args.CustomerLabels
	}

	// Update the protect surface
	updatedPS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, updatePS)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert the updated protect surface to JSON
	jsonData, err := json.Marshal(updatedPS)
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

// Delete handles protect surface deletion
func (t *ProtectSurfaceTools) Delete(ctx context.Context, req *mcp.CallToolRequest, args types.ProtectSurfaceParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]string, 0, len(args.IDs)+1)
	if args.ID != "" {
		ids = append(ids, args.ID)
	}
	ids = append(ids, args.IDs...)

	// Filter out empty values while preserving order
	filtered := ids[:0]
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	if len(filtered) == 0 {
		return nil, nil, fmt.Errorf("provide at least one protect surface id via 'id' or 'ids'")
	}

	var deleted []string
	var failed []string

	for _, id := range filtered {
		if err := auxoClient.ZeroTrust.DeleteProtectSurfaceByID(ctx, id); err != nil {
			friendly := client.FriendlyError(err)
			failed = append(failed, fmt.Sprintf("%s: %s", id, friendly.Error()))
			continue
		}
		deleted = append(deleted, id)
	}

	if len(failed) > 0 {
		var builder strings.Builder
		if len(deleted) > 0 {
			builder.WriteString("Deleted protect surfaces: ")
			builder.WriteString(strings.Join(deleted, ", "))
			builder.WriteString(". ")
		}
		builder.WriteString("Failed to delete: ")
		builder.WriteString(strings.Join(failed, "; "))
		return nil, nil, errors.New(builder.String())
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: "Protect surfaces deleted: " + strings.Join(deleted, ", "),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}
