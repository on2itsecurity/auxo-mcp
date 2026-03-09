package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
	"github.com/on2itsecurity/go-auxo/v2/zerotrust"
)

// ProtectSurfaceMeasureTools contains all protect surface measure management tools
type ProtectSurfaceMeasureTools struct {
	clientManager *client.Manager
}

// NewProtectSurfaceMeasureTools creates a new protect surface measure tools instance
func NewProtectSurfaceMeasureTools(clientManager *client.Manager) *ProtectSurfaceMeasureTools {
	return &ProtectSurfaceMeasureTools{clientManager: clientManager}
}

// validateMeasureUpdateParams validates conditional requirements for measure updates
func validateMeasureUpdateParams(args types.ProtectSurfaceMeasureParams) error {
	// Check assignment requirements
	if args.Assigned != nil && *args.Assigned && args.AssignmentPersonEmail == nil {
		return fmt.Errorf("assignment_person_email is required when assigned=true")
	}

	// Check implementation requirements
	if args.Implemented != nil && *args.Implemented && args.ImplementationPersonEmail == nil {
		return fmt.Errorf("implementation_person_email is required when implemented=true")
	}

	// Check evidence requirements
	if args.Evidenced != nil && *args.Evidenced && args.EvidencePersonEmail == nil {
		return fmt.Errorf("evidence_person_email is required when evidenced=true")
	}

	// Check risk acceptance requirements
	riskAcceptanceProvided := (args.RiskNoImplementationAccepted != nil && *args.RiskNoImplementationAccepted) ||
		(args.RiskNoEvidenceAccepted != nil && *args.RiskNoEvidenceAccepted)

	if riskAcceptanceProvided {
		if args.RiskAcceptancePersonEmail == nil {
			return fmt.Errorf("risk_acceptance_person_email is required when any risk acceptance field is true")
		}
		if args.RiskAcceptedComment == nil {
			return fmt.Errorf("risk_accepted_comment is required when any risk acceptance field is true")
		}
	}

	return nil
}

// List handles listing measures assigned to a protect surface
func (t *ProtectSurfaceMeasureTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.ProtectSurfaceMeasureParams) (*mcp.CallToolResult, any, error) {
	if args.ProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("protect_surface_id is required")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the protect surface
	protectSurface, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.ProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protect surface: %w", client.FriendlyError(err))
	}

	// Get all measures from catalog for reference
	measureGroups, err := auxoClient.ZeroTrust.GetMeasures(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get measures catalog: %w", client.FriendlyError(err))
	}

	// Create a lookup map for measure details
	measureLookup := make(map[string]map[string]interface{})
	for _, group := range measureGroups.Groups {
		for _, measure := range group.Measures {
			measureLookup[measure.Name] = map[string]interface{}{
				"caption":       measure.Caption,
				"explanation":   measure.Explanation,
				"mappings":      measure.Mappings,
				"group_name":    group.Name,
				"group_caption": group.Caption,
			}
		}
	}

	// Process assigned measures
	var result []map[string]interface{}
	for measureName, measureState := range protectSurface.Measures {
		measureInfo := map[string]interface{}{
			"measure_name":                    measureName,
			"protect_surface_id":              args.ProtectSurfaceID,
			"assigned":                        false,
			"implemented":                     false,
			"evidenced":                       false,
			"assignment_person":               "",
			"assignment_timestamp":            0,
			"implementation_person":           "",
			"implementation_timestamp":        0,
			"evidence_person":                 "",
			"evidence_timestamp":              0,
			"risk_no_implementation_accepted": false,
			"risk_no_evidence_accepted":       false,
			"risk_accepted_comment":           "",
			"risk_acceptance_person":          "",
			"risk_acceptance_timestamp":       0,
		}

		// Add catalog information if available
		if catalogInfo, exists := measureLookup[measureName]; exists {
			for key, value := range catalogInfo {
				measureInfo[key] = value
			}
		}

		// Add assignment information
		if measureState.Assignment != nil {
			measureInfo["assigned"] = measureState.Assignment.Assigned
			measureInfo["assignment_person"] = measureState.Assignment.LastDeterminedByPersonID
			measureInfo["assignment_timestamp"] = measureState.Assignment.LastDeterminedTimestamp
		}

		// Add implementation information
		if measureState.Implementation != nil {
			measureInfo["implemented"] = measureState.Implementation.Implemented
			measureInfo["implementation_person"] = measureState.Implementation.LastDeterminedByPersonID
			measureInfo["implementation_timestamp"] = measureState.Implementation.LastDeterminedTimestamp
		}

		// Add evidence information
		if measureState.Evidence != nil {
			measureInfo["evidenced"] = measureState.Evidence.Evidenced
			measureInfo["evidence_person"] = measureState.Evidence.LastDeterminedByPersonID
			measureInfo["evidence_timestamp"] = measureState.Evidence.LastDeterminedTimestamp
		}

		// Add risk acceptance information
		if measureState.RiskAcceptance != nil {
			measureInfo["risk_no_implementation_accepted"] = measureState.RiskAcceptance.RiskNoImplementationAccepted
			measureInfo["risk_no_evidence_accepted"] = measureState.RiskAcceptance.RiskNoEvidenceAccepted
			measureInfo["risk_accepted_comment"] = measureState.RiskAcceptance.RiskAcceptedComment
			measureInfo["risk_acceptance_person"] = measureState.RiskAcceptance.LastDeterminedByPersonID
			measureInfo["risk_acceptance_timestamp"] = measureState.RiskAcceptance.LastDeterminedTimestamp
		}

		// Apply filters if provided
		if args.MeasureName != "" && measureName != args.MeasureName {
			continue
		}

		if args.AssignedOnly != nil && *args.AssignedOnly {
			if !measureInfo["assigned"].(bool) {
				continue
			}
		}

		if args.ImplementedOnly != nil && *args.ImplementedOnly {
			if !measureInfo["implemented"].(bool) {
				continue
			}
		}

		result = append(result, measureInfo)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(result)
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

// Update handles updating any aspect of a measure assigned to a protect surface
// This function supports partial updates - only provided fields will be updated
func (t *ProtectSurfaceMeasureTools) Update(ctx context.Context, req *mcp.CallToolRequest, args types.ProtectSurfaceMeasureParams) (*mcp.CallToolResult, any, error) {
	if args.ProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("protect_surface_id is required")
	}
	if args.MeasureName == "" {
		return nil, nil, fmt.Errorf("measure_name is required")
	}

	// Validate conditional requirements
	if err := validateMeasureUpdateParams(args); err != nil {
		return nil, nil, err
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the protect surface
	protectSurface, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.ProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protect surface: %w", client.FriendlyError(err))
	}

	// Initialize measures map if nil
	if protectSurface.Measures == nil {
		protectSurface.Measures = make(map[string]zerotrust.MeasureState)
	}

	// Get existing measure state or create new one
	measureState := protectSurface.Measures[args.MeasureName]
	currentTime := int(time.Now().Unix())

	// Update assignment if provided
	if args.Assigned != nil {
		if measureState.Assignment == nil {
			measureState.Assignment = &zerotrust.Assignment{}
		}
		measureState.Assignment.Assigned = *args.Assigned
		measureState.Assignment.LastDeterminedTimestamp = currentTime

		// Set assignment person email (required for assignment changes)
		if args.AssignmentPersonEmail != nil {
			measureState.Assignment.LastDeterminedByPersonID = *args.AssignmentPersonEmail
		} else {
			return nil, nil, fmt.Errorf("assignment_person_email is required when updating assignment status")
		}
	}

	// Update implementation if provided
	if args.Implemented != nil {
		if measureState.Implementation == nil {
			measureState.Implementation = &zerotrust.Implementation{}
		}
		measureState.Implementation.Implemented = *args.Implemented
		measureState.Implementation.LastDeterminedTimestamp = currentTime

		// Set implementation person email (required for implementation changes)
		if args.ImplementationPersonEmail != nil {
			measureState.Implementation.LastDeterminedByPersonID = *args.ImplementationPersonEmail
		} else {
			return nil, nil, fmt.Errorf("implementation_person_email is required when updating implementation status")
		}
	}

	// Update evidence if provided
	if args.Evidenced != nil {
		if measureState.Evidence == nil {
			measureState.Evidence = &zerotrust.Evidence{}
		}
		measureState.Evidence.Evidenced = *args.Evidenced
		measureState.Evidence.LastDeterminedTimestamp = currentTime

		// Set evidence person ID (required for evidence changes)
		if args.EvidencePersonEmail != nil {
			measureState.Evidence.LastDeterminedByPersonID = *args.EvidencePersonEmail
		} else {
			return nil, nil, fmt.Errorf("evidence_person_email is required when updating evidence status")
		}
	}

	// Update risk acceptance if any risk acceptance fields are provided
	if args.RiskNoImplementationAccepted != nil || args.RiskNoEvidenceAccepted != nil || args.RiskAcceptedComment != nil {
		if measureState.RiskAcceptance == nil {
			measureState.RiskAcceptance = &zerotrust.RiskAcceptance{}
		}

		// Set risk acceptance person email (required for risk acceptance changes)
		if args.RiskAcceptancePersonEmail != nil {
			measureState.RiskAcceptance.LastDeterminedByPersonID = *args.RiskAcceptancePersonEmail
		} else {
			return nil, nil, fmt.Errorf("risk_acceptance_person_email is required when updating risk acceptance")
		}

		measureState.RiskAcceptance.LastDeterminedTimestamp = currentTime

		// Update specific risk acceptance fields if provided
		if args.RiskNoImplementationAccepted != nil {
			measureState.RiskAcceptance.RiskNoImplementationAccepted = *args.RiskNoImplementationAccepted
		}
		if args.RiskNoEvidenceAccepted != nil {
			measureState.RiskAcceptance.RiskNoEvidenceAccepted = *args.RiskNoEvidenceAccepted
		}
		if args.RiskAcceptedComment != nil {
			measureState.RiskAcceptance.RiskAcceptedComment = *args.RiskAcceptedComment
		}
	}

	// Update the measure state in the protect surface
	protectSurface.Measures[args.MeasureName] = measureState

	// Update the protect surface
	updatedPS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *protectSurface)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update protect surface: %w", client.FriendlyError(err))
	}

	// Build comprehensive response with current state
	response := map[string]interface{}{
		"protect_surface_id": updatedPS.ID,
		"measure_name":       args.MeasureName,
		"success":            true,
		"updated_fields":     []string{},
	}

	// Add current state information
	if measureState.Assignment != nil {
		response["assigned"] = measureState.Assignment.Assigned
		response["assignment_person"] = measureState.Assignment.LastDeterminedByPersonID
		response["assignment_timestamp"] = measureState.Assignment.LastDeterminedTimestamp
		if args.Assigned != nil {
			response["updated_fields"] = append(response["updated_fields"].([]string), "assignment")
		}
	}

	if measureState.Implementation != nil {
		response["implemented"] = measureState.Implementation.Implemented
		response["implementation_person"] = measureState.Implementation.LastDeterminedByPersonID
		response["implementation_timestamp"] = measureState.Implementation.LastDeterminedTimestamp
		if args.Implemented != nil {
			response["updated_fields"] = append(response["updated_fields"].([]string), "implementation")
		}
	}

	if measureState.Evidence != nil {
		response["evidenced"] = measureState.Evidence.Evidenced
		response["evidence_person"] = measureState.Evidence.LastDeterminedByPersonID
		response["evidence_timestamp"] = measureState.Evidence.LastDeterminedTimestamp
		if args.Evidenced != nil {
			response["updated_fields"] = append(response["updated_fields"].([]string), "evidence")
		}
	}

	if measureState.RiskAcceptance != nil {
		response["risk_no_implementation_accepted"] = measureState.RiskAcceptance.RiskNoImplementationAccepted
		response["risk_no_evidence_accepted"] = measureState.RiskAcceptance.RiskNoEvidenceAccepted
		response["risk_accepted_comment"] = measureState.RiskAcceptance.RiskAcceptedComment
		response["risk_acceptance_person"] = measureState.RiskAcceptance.LastDeterminedByPersonID
		response["risk_acceptance_timestamp"] = measureState.RiskAcceptance.LastDeterminedTimestamp
		if args.RiskNoImplementationAccepted != nil || args.RiskNoEvidenceAccepted != nil || args.RiskAcceptedComment != nil {
			response["updated_fields"] = append(response["updated_fields"].([]string), "risk_acceptance")
		}
	}

	jsonData, err := json.Marshal(response)
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

// Remove handles removing a measure assignment from a protect surface
func (t *ProtectSurfaceMeasureTools) Remove(ctx context.Context, req *mcp.CallToolRequest, args types.ProtectSurfaceMeasureParams) (*mcp.CallToolResult, any, error) {
	if args.ProtectSurfaceID == "" {
		return nil, nil, fmt.Errorf("protect_surface_id is required")
	}
	if args.MeasureName == "" {
		return nil, nil, fmt.Errorf("measure_name is required")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get the protect surface
	protectSurface, err := auxoClient.ZeroTrust.GetProtectSurfaceByID(ctx, args.ProtectSurfaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protect surface: %w", client.FriendlyError(err))
	}

	// Check if measure exists
	_, exists := protectSurface.Measures[args.MeasureName]
	if !exists {
		return nil, nil, fmt.Errorf("measure %s is not assigned to protect surface %s", args.MeasureName, args.ProtectSurfaceID)
	}

	// Remove the measure
	delete(protectSurface.Measures, args.MeasureName)

	// Update the protect surface
	updatedPS, err := auxoClient.ZeroTrust.UpdateProtectSurface(ctx, *protectSurface)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update protect surface: %w", client.FriendlyError(err))
	}

	// Return success response
	response := map[string]interface{}{
		"protect_surface_id": updatedPS.ID,
		"measure_name":       args.MeasureName,
		"removed":            true,
		"success":            true,
	}

	jsonData, err := json.Marshal(response)
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
