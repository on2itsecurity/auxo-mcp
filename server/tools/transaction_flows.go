package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
	"github.com/on2itsecurity/go-auxo/v2/zerotrust"
)

// TransactionFlowTools contains all transaction flow related tools
type TransactionFlowTools struct {
	clientManager *client.Manager
}

// NewTransactionFlowTools creates a new transaction flow tools instance
func NewTransactionFlowTools(clientManager *client.Manager) *TransactionFlowTools {
	return &TransactionFlowTools{clientManager: clientManager}
}

// CreateFlow handles flow creation between protect surfaces with mutual consensus
func (t *TransactionFlowTools) CreateFlow(ctx context.Context, req *mcp.CallToolRequest, args types.TransactionFlowParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.SourceProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("source_protect_surface_id is required")
	}
	if args.DestinationProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("destination_protect_surface_id is required")
	}
	if args.Allow == nil {
		return nil, nil, fmt.Errorf("allow is required")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get both protect surfaces
	sourcePS, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.SourceProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get source protect surface: %w", client.FriendlyError(err))
	}

	destinationPS, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.DestinationProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get destination protect surface: %w", client.FriendlyError(err))
	}

	// Create flow object
	flow := zerotrust.Flow{
		Allow: args.Allow,
	}

	// Initialize flow maps if they don't exist
	if sourcePS.FlowsToOtherPS == nil {
		sourcePS.FlowsToOtherPS = make(map[string]zerotrust.Flow)
	}
	if destinationPS.FlowsFromOtherPS == nil {
		destinationPS.FlowsFromOtherPS = make(map[string]zerotrust.Flow)
	}

	// Set mutual consensus flows
	sourcePS.FlowsToOtherPS[args.DestinationProtectSurfaceID] = flow
	destinationPS.FlowsFromOtherPS[args.SourceProtectSurfaceID] = flow

	// Update both protect surfaces
	updatedSourcePS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *sourcePS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update source protect surface: %w", client.FriendlyError(err))
	}

	updatedDestinationPS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *destinationPS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update destination protect surface: %w", client.FriendlyError(err))
	}

	// Create response with both updated protect surfaces
	response := map[string]interface{}{
		"source_protect_surface":      updatedSourcePS,
		"destination_protect_surface": updatedDestinationPS,
		"flow_created": map[string]interface{}{
			"from":  args.SourceProtectSurfaceID,
			"to":    args.DestinationProtectSurfaceID,
			"allow": *args.Allow,
		},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// CreateExternalFlow handles flows to/from outside the organization
func (t *TransactionFlowTools) CreateExternalFlow(ctx context.Context, req *mcp.CallToolRequest, args types.ExternalFlowParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("protect_surface_id is required")
	}
	if args.Direction == "" {
		return nil, nil, fmt.Errorf("direction is required (inbound or outbound)")
	}
	if args.Allow == nil {
		return nil, nil, fmt.Errorf("allow is required")
	}

	// Validate direction
	if args.Direction != "inbound" && args.Direction != "outbound" {
		return nil, nil, fmt.Errorf("direction must be either 'inbound' or 'outbound'")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the protect surface
	ps, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.ProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protect surface: %w", client.FriendlyError(err))
	}

	// Create flow object
	flow := zerotrust.Flow{
		Allow: args.Allow,
	}

	// Set the appropriate flow based on direction
	if args.Direction == "inbound" {
		ps.FlowsFromOutside = flow
	} else {
		ps.FlowsToOutside = flow
	}

	// Update the protect surface
	updatedPS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *ps)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update protect surface: %w", client.FriendlyError(err))
	}

	// Create response
	response := map[string]interface{}{
		"protect_surface": updatedPS,
		"external_flow_created": map[string]interface{}{
			"protect_surface_id": args.ProtectSurfaceID,
			"direction":          args.Direction,
			"allow":              *args.Allow,
		},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// ListFlows handles listing all flows for a specific protect surface
func (t *TransactionFlowTools) ListFlows(ctx context.Context, req *mcp.CallToolRequest, args types.FlowListParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("protect_surface_id is required")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the protect surface
	ps, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.ProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protect surface: %w", client.FriendlyError(err))
	}

	// Create response with all flow information
	response := map[string]interface{}{
		"protect_surface_id":   args.ProtectSurfaceID,
		"protect_surface_name": ps.Name,
		"flows_to_other_ps":    ps.FlowsToOtherPS,
		"flows_from_other_ps":  ps.FlowsFromOtherPS,
		"flows_from_outside":   ps.FlowsFromOutside,
		"flows_to_outside":     ps.FlowsToOutside,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// DeleteFlow handles flow deletion with mutual consensus
func (t *TransactionFlowTools) DeleteFlow(ctx context.Context, req *mcp.CallToolRequest, args types.TransactionFlowParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.SourceProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("source_protect_surface_id is required")
	}
	if args.DestinationProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("destination_protect_surface_id is required")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get both protect surfaces
	sourcePS, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.SourceProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get source protect surface: %w", client.FriendlyError(err))
	}

	destinationPS, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.DestinationProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get destination protect surface: %w", client.FriendlyError(err))
	}

	// Check if flows exist
	flowExists := false
	if sourcePS.FlowsToOtherPS != nil {
		if _, exists := sourcePS.FlowsToOtherPS[args.DestinationProtectSurfaceID]; exists {
			delete(sourcePS.FlowsToOtherPS, args.DestinationProtectSurfaceID)
			flowExists = true
		}
	}
	if destinationPS.FlowsFromOtherPS != nil {
		if _, exists := destinationPS.FlowsFromOtherPS[args.SourceProtectSurfaceID]; exists {
			delete(destinationPS.FlowsFromOtherPS, args.SourceProtectSurfaceID)
			flowExists = true
		}
	}

	if !flowExists {
		return nil, nil, fmt.Errorf("no flow found between protect surfaces %s and %s", args.SourceProtectSurfaceID, args.DestinationProtectSurfaceID)
	}

	// Update both protect surfaces
	updatedSourcePS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *sourcePS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update source protect surface: %w", client.FriendlyError(err))
	}

	updatedDestinationPS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *destinationPS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update destination protect surface: %w", client.FriendlyError(err))
	}

	// Create response
	response := map[string]interface{}{
		"source_protect_surface":      updatedSourcePS,
		"destination_protect_surface": updatedDestinationPS,
		"flow_deleted": map[string]interface{}{
			"from": args.SourceProtectSurfaceID,
			"to":   args.DestinationProtectSurfaceID,
		},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// DeleteExternalFlow handles deletion of external flows
func (t *TransactionFlowTools) DeleteExternalFlow(ctx context.Context, req *mcp.CallToolRequest, args types.ExternalFlowParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("protect_surface_id is required")
	}
	if args.Direction == "" {
		return nil, nil, fmt.Errorf("direction is required (inbound or outbound)")
	}

	// Validate direction
	if args.Direction != "inbound" && args.Direction != "outbound" {
		return nil, nil, fmt.Errorf("direction must be either 'inbound' or 'outbound'")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the protect surface
	ps, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.ProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protect surface: %w", client.FriendlyError(err))
	}

	// Clear the appropriate flow based on direction
	if args.Direction == "inbound" {
		ps.FlowsFromOutside = zerotrust.Flow{}
	} else {
		ps.FlowsToOutside = zerotrust.Flow{}
	}

	// Update the protect surface
	updatedPS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *ps)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update protect surface: %w", client.FriendlyError(err))
	}

	// Create response
	response := map[string]interface{}{
		"protect_surface": updatedPS,
		"external_flow_deleted": map[string]interface{}{
			"protect_surface_id": args.ProtectSurfaceID,
			"direction":          args.Direction,
		},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		},
	}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}
