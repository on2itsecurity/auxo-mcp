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

// LocationTools contains all location related tools
type LocationTools struct {
	clientManager *client.Manager
}

// NewLocationTools creates a new location tools instance
func NewLocationTools(clientManager *client.Manager) *LocationTools {
	return &LocationTools{clientManager: clientManager}
}

// Create handles location creation
func (t *LocationTools) Create(ctx context.Context, req *mcp.CallToolRequest, args types.LocationParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for creation
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name is required for creating a location")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Build the location object with all provided fields
	location := zerotrust.Location{
		Name:          args.Name,
		UniquenessKey: args.UniquenessKey,
	}

	// Handle coordinates if provided
	if args.Latitude != nil || args.Longitude != nil {
		location.Coords = zerotrust.Coords{
			Latitude:  0,
			Longitude: 0,
		}
		if args.Latitude != nil {
			location.Coords.Latitude = *args.Latitude
		}
		if args.Longitude != nil {
			location.Coords.Longitude = *args.Longitude
		}
	}

	loc, err := auxoClient.ZeroTrust.CreateLocationByObject(ctx, location, false)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert the created location to JSON
	jsonData, err := json.Marshal(loc)
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

// List handles location listing
func (t *LocationTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	allLocations, err := auxoClient.ZeroTrust.GetLocations(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allLocations)
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

// Update handles location updates
func (t *LocationTools) Update(ctx context.Context, req *mcp.CallToolRequest, args types.LocationParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for update
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for updating a location")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the current location to merge with updates
	currentLoc, err := auxoClient.ZeroTrust.GetLocationByID(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Update fields if provided, otherwise keep existing values
	updateLoc := *currentLoc

	// Update string fields if provided
	if args.Name != "" {
		updateLoc.Name = args.Name
	}
	if args.UniquenessKey != "" {
		updateLoc.UniquenessKey = args.UniquenessKey
	}

	// Update coordinates if provided
	if args.Latitude != nil {
		updateLoc.Coords.Latitude = *args.Latitude
	}
	if args.Longitude != nil {
		updateLoc.Coords.Longitude = *args.Longitude
	}

	// Update the location
	updatedLoc, err := auxoClient.ZeroTrust.UpdateLocation(ctx, updateLoc)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert the updated location to JSON
	jsonData, err := json.Marshal(updatedLoc)
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

// Delete handles location deletion
func (t *LocationTools) Delete(ctx context.Context, req *mcp.CallToolRequest, args types.LocationParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]string, 0, len(args.IDs)+1)
	if args.ID != "" {
		ids = append(ids, args.ID)
	}
	ids = append(ids, args.IDs...)

	filtered := ids[:0]
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	if len(filtered) == 0 {
		return nil, nil, fmt.Errorf("provide at least one location id via 'id' or 'ids'")
	}

	var deleted []string
	var failed []string

	for _, id := range filtered {
		if err := auxoClient.ZeroTrust.DeleteLocationByID(ctx, id); err != nil {
			friendly := client.FriendlyError(err)
			failed = append(failed, fmt.Sprintf("%s: %s", id, friendly.Error()))
			continue
		}
		deleted = append(deleted, id)
	}

	if len(failed) > 0 {
		var builder strings.Builder
		if len(deleted) > 0 {
			builder.WriteString("Deleted locations: ")
			builder.WriteString(strings.Join(deleted, ", "))
			builder.WriteString(". ")
		}
		builder.WriteString("Failed to delete: ")
		builder.WriteString(strings.Join(failed, "; "))
		return nil, nil, errors.New(builder.String())
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: "Locations deleted: " + strings.Join(deleted, ", "),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}
