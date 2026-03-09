package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
	"github.com/on2itsecurity/go-auxo/v2/caseintegration"
)

// CaseTools contains all case/ticket related tools
type CaseTools struct {
	clientManager *client.Manager
}

// NewCaseTools creates a new case tools instance
func NewCaseTools(clientManager *client.Manager) *CaseTools {
	return &CaseTools{clientManager: clientManager}
}

// Create handles case creation
func (t *CaseTools) Create(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for creation
	if args.Subject == "" {
		return nil, nil, fmt.Errorf("subject is required for creating a case")
	}
	if args.Note == "" {
		return nil, nil, fmt.Errorf("note is required for creating a case")
	}
	if args.Priority == nil {
		return nil, nil, fmt.Errorf("priority is required for creating a case")
	}
	if args.CaseType == "" {
		return nil, nil, fmt.Errorf("case_type is required for creating a case (valid values: 'securityincident', 'incident', 'change', 'standardchange', 'inforequest')")
	}
	if args.PrimaryContactEmail == "" {
		return nil, nil, fmt.Errorf("primary_contact_email is required for creating a case")
	}

	// Validate priority range
	if *args.Priority < 1 || *args.Priority > 4 {
		return nil, nil, fmt.Errorf("priority must be between 1 and 4, where 1 is highest priority")
	}

	// Validate case type
	validCaseTypes := map[string]bool{
		"securityincident": true,
		"incident":         true,
		"change":           true,
		"standardchange":   true,
		"inforequest":      true,
	}
	if !validCaseTypes[args.CaseType] {
		return nil, nil, fmt.Errorf("invalid case_type '%s'. Valid values are: 'securityincident', 'incident', 'change', 'standardchange', 'inforequest'", args.CaseType)
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Build the case object
	newCase := caseintegration.Case{
		ID:                  args.ID, // Optional, will be generated if not provided
		Subject:             args.Subject,
		Note:                args.Note,
		Priority:            *args.Priority,
		CaseType:            args.CaseType,
		PrimaryContactEmail: args.PrimaryContactEmail,
	}

	err = auxoClient.CaseIntegration.CreateCaseByObject(ctx, newCase)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Case created successfully with subject: %s", args.Subject),
	}
	if args.ID != "" {
		successMsg["id"] = args.ID
	}

	jsonData, err := json.Marshal(successMsg)
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

// Get handles getting a case by ID
func (t *CaseTools) Get(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for getting a case")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	caseData, err := auxoClient.CaseIntegration.GetCaseByID(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Convert the case to JSON
	jsonData, err := json.Marshal(caseData)
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

// UpdatePriority handles updating the priority of a case
func (t *CaseTools) UpdatePriority(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for updating case priority")
	}
	if args.NewPriority == nil {
		return nil, nil, fmt.Errorf("new_priority is required for updating case priority")
	}

	// Validate priority range
	if *args.NewPriority < 1 || *args.NewPriority > 4 {
		return nil, nil, fmt.Errorf("priority must be between 1 and 4, where 1 is highest priority")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = auxoClient.CaseIntegration.UpdatePriorityOfCase(ctx, args.ID, *args.NewPriority)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Case priority updated successfully to %d", *args.NewPriority),
		"id":      args.ID,
	}

	jsonData, err := json.Marshal(successMsg)
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

// UpdatePrimaryContact handles updating the primary contact of a case
func (t *CaseTools) UpdatePrimaryContact(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for updating primary contact")
	}
	if args.NewPrimaryContactEmail == "" {
		return nil, nil, fmt.Errorf("new_primary_contact_email is required for updating primary contact")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = auxoClient.CaseIntegration.UpdatePrimaryContactOfCase(ctx, args.ID, args.NewPrimaryContactEmail)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Case primary contact updated successfully to %s", args.NewPrimaryContactEmail),
		"id":      args.ID,
	}

	jsonData, err := json.Marshal(successMsg)
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

// UpdateSubject handles updating the subject of a case
func (t *CaseTools) UpdateSubject(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for updating case subject")
	}
	if args.NewSubject == "" {
		return nil, nil, fmt.Errorf("new_subject is required for updating case subject")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = auxoClient.CaseIntegration.UpdateSubjectOfCase(ctx, args.ID, args.NewSubject)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Case subject updated successfully to: %s", args.NewSubject),
		"id":      args.ID,
	}

	jsonData, err := json.Marshal(successMsg)
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

// Escalate handles escalating a case
func (t *CaseTools) Escalate(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for escalating a case")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = auxoClient.CaseIntegration.EscalateCase(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": "Case escalated successfully",
		"id":      args.ID,
	}

	jsonData, err := json.Marshal(successMsg)
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

// Deescalate handles de-escalating a case
func (t *CaseTools) Deescalate(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for de-escalating a case")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = auxoClient.CaseIntegration.DeescalateCase(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": "Case de-escalated successfully",
		"id":      args.ID,
	}

	jsonData, err := json.Marshal(successMsg)
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

// AddNote handles adding a note to a case
func (t *CaseTools) AddNote(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for adding a note to a case")
	}
	if args.AdditionalNote == "" {
		return nil, nil, fmt.Errorf("additional_note is required for adding a note to a case")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = auxoClient.CaseIntegration.AddNoteToCase(ctx, args.ID, args.AdditionalNote)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": "Note added to case successfully",
		"id":      args.ID,
	}

	jsonData, err := json.Marshal(successMsg)
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

// Close handles requesting to close a case
func (t *CaseTools) Close(ctx context.Context, req *mcp.CallToolRequest, args types.CaseParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required for closing a case")
	}

	auxoClient, err := t.clientManager.CreateCaseClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = auxoClient.CaseIntegration.RequestCaseClose(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": "Case close requested successfully. The case will be reviewed by an engineer and closed if no further work is needed.",
		"id":      args.ID,
	}

	jsonData, err := json.Marshal(successMsg)
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
