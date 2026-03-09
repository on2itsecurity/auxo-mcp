package resources

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/tools"
	"github.com/on2itsecurity/go-auxo/v2/zerotrust"
)

// Handlers contains all resource handlers
type Handlers struct {
	clientManager *client.Manager
}

// NewHandlers creates a new resource handlers instance
func NewHandlers(clientManager *client.Manager) *Handlers {
	return &Handlers{clientManager: clientManager}
}

// ProtectSurfacesResource returns the protect surfaces resource definition
func (h *Handlers) ProtectSurfacesResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "auxo://protect-surfaces",
		Name:        "protect-surfaces",
		Title:       "All Protect Surfaces",
		Description: "Retrieve all protect surfaces for the customer",
		MIMEType:    "application/json",
	}
}

// ProtectSurfaces handles protect surfaces resource requests
func (h *Handlers) ProtectSurfaces(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	auxoClient, err := h.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	allProtectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces(ctx)
	if err != nil {
		return nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allProtectSurfaces)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// ContactsResource returns the contacts resource definition
func (h *Handlers) ContactsResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "auxo://contacts",
		Name:        "contacts",
		Title:       "All Contacts",
		Description: "Retrieve all contacts for the customer",
		MIMEType:    "application/json",
	}
}

// Contacts handles contacts resource requests
func (h *Handlers) Contacts(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	auxoClient, err := h.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	allContacts, err := auxoClient.CRM.GetContacts(ctx)
	if err != nil {
		return nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allContacts)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// LocationsResource returns the locations resource definition
func (h *Handlers) LocationsResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "auxo://locations",
		Name:        "locations",
		Title:       "All Locations",
		Description: "Retrieve all locations from the ZeroTrust system",
		MIMEType:    "application/json",
	}
}

// Locations handles locations resource requests
func (h *Handlers) Locations(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	auxoClient, err := h.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	allLocations, err := auxoClient.ZeroTrust.GetLocations(ctx)
	if err != nil {
		return nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allLocations)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// StatesResource returns the states resource definition
func (h *Handlers) StatesResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "auxo://states",
		Name:        "states",
		Title:       "All States",
		Description: "Retrieve all states from the ZeroTrust system",
		MIMEType:    "application/json",
	}
}

// States handles states resource requests
func (h *Handlers) States(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	auxoClient, err := h.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	allStates, err := auxoClient.ZeroTrust.GetStates(ctx)
	if err != nil {
		return nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allStates)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// AssetsResource returns the assets resource definition
func (h *Handlers) AssetsResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "auxo://assets",
		Name:        "assets",
		Title:       "All Assets",
		Description: "Retrieve all assets from the Asset management system",
		MIMEType:    "application/json",
	}
}

// Assets handles assets resource requests
func (h *Handlers) Assets(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	auxoClient, err := h.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	allAssets, err := auxoClient.Asset.GetAssets(ctx)
	if err != nil {
		return nil, client.FriendlyError(err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allAssets)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// TransactionFlowsResource returns the transaction flows resource definition
func (h *Handlers) TransactionFlowsResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "auxo://transaction-flows",
		Name:        "transaction-flows",
		Title:       "All Transaction Flows",
		Description: "Retrieve all transaction flows between protect surfaces in the zero trust network",
		MIMEType:    "application/json",
	}
}

// TransactionFlows handles transaction flows resource requests
func (h *Handlers) TransactionFlows(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	auxoClient, err := h.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	// Get all protect surfaces to extract flow information
	allProtectSurfaces, err := auxoClient.ZeroTrust.GetProtectSurfaces(ctx)
	if err != nil {
		return nil, client.FriendlyError(err)
	}

	// Create a comprehensive flow map
	var allFlows []map[string]interface{}

	for _, ps := range allProtectSurfaces {
		// Internal flows (to other protect surfaces)
		for destID, flow := range ps.FlowsToOtherPS {
			flowData := map[string]interface{}{
				"flow_type":                        "internal",
				"source_protect_surface_id":        ps.ID,
				"source_protect_surface_name":      ps.Name,
				"destination_protect_surface_id":   destID,
				"destination_protect_surface_name": "", // Will be filled if available
				"allow":                            flow.Allow,
				"direction":                        "outbound",
			}
			allFlows = append(allFlows, flowData)
		}

		// External flows (to/from outside)
		// Check if flows to outside are configured (non-zero struct)
		if ps.FlowsToOutside != (zerotrust.Flow{}) {
			flowData := map[string]interface{}{
				"flow_type":                   "external",
				"source_protect_surface_id":   ps.ID,
				"source_protect_surface_name": ps.Name,
				"destination":                 "outside",
				"allow":                       ps.FlowsToOutside.Allow,
				"direction":                   "outbound",
			}
			allFlows = append(allFlows, flowData)
		}

		if ps.FlowsFromOutside != (zerotrust.Flow{}) {
			flowData := map[string]interface{}{
				"flow_type":                        "external",
				"destination_protect_surface_id":   ps.ID,
				"destination_protect_surface_name": ps.Name,
				"source":                           "outside",
				"allow":                            ps.FlowsFromOutside.Allow,
				"direction":                        "inbound",
			}
			allFlows = append(allFlows, flowData)
		}
	}

	// Fill in destination protect surface names for internal flows
	psNameMap := make(map[string]string)
	for _, ps := range allProtectSurfaces {
		psNameMap[ps.ID] = ps.Name
	}

	for _, flow := range allFlows {
		if destID, exists := flow["destination_protect_surface_id"]; exists {
			if name, found := psNameMap[destID.(string)]; found {
				flow["destination_protect_surface_name"] = name
			}
		}
	}

	// Convert to JSON
	jsonData, err := json.Marshal(allFlows)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

// MeasuresResource returns the measures resource definition
func (h *Handlers) MeasuresResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "auxo://measures",
		Name:        "measures",
		Title:       "All Security Measures",
		Description: "Retrieve all available security measures from the AUXO catalog",
		MIMEType:    "application/json",
	}
}

// Measures handles measures resource requests
func (h *Handlers) Measures(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	auxoClient, err := h.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, err
	}

	// Get all available measures from the catalog
	measureGroups, err := auxoClient.ZeroTrust.GetMeasures(ctx)
	if err != nil {
		return nil, client.FriendlyError(err)
	}

	// Convert to response format - flatten all measures from all groups
	allMeasures := tools.FlattenMeasureGroups(*measureGroups)

	// Convert to JSON
	jsonData, err := json.Marshal(allMeasures)
	if err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}
