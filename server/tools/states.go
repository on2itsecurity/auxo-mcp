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

// StateTools contains all state related tools
type StateTools struct {
	clientManager *client.Manager
}

// NewStateTools creates a new state tools instance
func NewStateTools(clientManager *client.Manager) *StateTools {
	return &StateTools{clientManager: clientManager}
}

// Create handles state creation
func (t *StateTools) Create(ctx context.Context, req *mcp.CallToolRequest, args types.StateParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for creation
	if args.ProtectSurface == "" {
		return nil, nil, fmt.Errorf("protectsurface_id is required for creating a state")
	}
	if args.Location == "" {
		return nil, nil, fmt.Errorf("location_id is required for creating a state")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Build the state object with all provided fields
	state := zerotrust.State{
		UniquenessKey:    args.UniquenessKey,
		Description:      args.Description,
		ProtectSurface:   args.ProtectSurface,
		ContentType:      args.ContentType,
		Location:         args.Location,
		ExistsOnAssetIDs: args.ExistsOnAssetIDs,
		Maintainer:       args.Maintainer,
	}

	// Handle content if provided (convert to pointer)
	if len(args.Content) > 0 {
		state.Content = &args.Content
	}

	// Set default content type if not provided
	if state.ContentType == "" {
		state.ContentType = "ipv4"
	}

	st, err := auxoClient.ZeroTrust.CreateStateByObject(ctx, state)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert the created state to JSON
	jsonData, err := json.Marshal(st)
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

// List handles state listing
func (t *StateTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	allStates, err := auxoClient.ZeroTrust.GetStates(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allStates)
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

// Update handles state updates
func (t *StateTools) Update(ctx context.Context, req *mcp.CallToolRequest, args types.StateParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for update
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for updating a state")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the current state to merge with updates
	currentState, err := auxoClient.ZeroTrust.GetStateByID(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Update fields if provided, otherwise keep existing values
	updateState := *currentState

	// Update string fields if provided
	if args.UniquenessKey != "" {
		updateState.UniquenessKey = args.UniquenessKey
	}
	if args.Description != "" {
		updateState.Description = args.Description
	}
	if args.ProtectSurface != "" {
		updateState.ProtectSurface = args.ProtectSurface
	}
	if args.ContentType != "" {
		updateState.ContentType = args.ContentType
	}
	if args.Location != "" {
		updateState.Location = args.Location
	}
	if args.Maintainer != "" {
		updateState.Maintainer = args.Maintainer
	}

	// Update slice fields if provided
	if args.ExistsOnAssetIDs != nil {
		updateState.ExistsOnAssetIDs = args.ExistsOnAssetIDs
	}
	if args.Content != nil {
		updateState.Content = &args.Content
	}

	// Update the state
	updatedState, err := auxoClient.ZeroTrust.UpdateState(ctx, updateState)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert the updated state to JSON
	jsonData, err := json.Marshal(updatedState)
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

// Delete handles state deletion
func (t *StateTools) Delete(ctx context.Context, req *mcp.CallToolRequest, args types.StateParams) (*mcp.CallToolResult, any, error) {
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
		return nil, nil, fmt.Errorf("provide at least one state id via 'id' or 'ids'")
	}

	var deleted []string
	var failed []string

	for _, id := range filtered {
		if err := auxoClient.ZeroTrust.DeleteStateByID(ctx, id); err != nil {
			friendly := client.FriendlyError(err)
			failed = append(failed, fmt.Sprintf("%s: %s", id, friendly.Error()))
			continue
		}
		deleted = append(deleted, id)
	}

	if len(failed) > 0 {
		var builder strings.Builder
		if len(deleted) > 0 {
			builder.WriteString("Deleted states: ")
			builder.WriteString(strings.Join(deleted, ", "))
			builder.WriteString(". ")
		}
		builder.WriteString("Failed to delete: ")
		builder.WriteString(strings.Join(failed, "; "))
		return nil, nil, errors.New(builder.String())
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: "States deleted: " + strings.Join(deleted, ", "),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}
