package types

// EmptyParams for operations that don't require input
type EmptyParams struct {
	_ struct{} `json:"-" jsonschema:"title=No Parameters Required,description=This operation does not require any input parameters"`
}

// ProtectSurfaceParams for protect surface operations (create, update, and delete)
type ProtectSurfaceParams struct {
	ID                      string            `json:"id,omitempty" jsonschema:"The ID of the protect surface (required for update and delete operations)"`
	IDs                     []string          `json:"ids,omitempty" jsonschema:"Optional list of protect surface IDs for bulk delete operations"`
	Name                    string            `json:"name,omitempty" jsonschema:"The name of the protect surface (required for create operations)"`
	Description             string            `json:"description,omitempty" jsonschema:"The description of the protect surface"`
	UniquenessKey           string            `json:"uniqueness_key,omitempty" jsonschema:"Key to prevent duplicates, especially in parallel processing"`
	MainContactPersonID     string            `json:"main_contact_person_id,omitempty" jsonschema:"Contact Person ID, must be a valid ID within the relation"`
	SecurityContactPersonID string            `json:"security_contact_person_id,omitempty" jsonschema:"Security Person ID, must be a valid ID within the relation"`
	InControlBoundary       *bool             `json:"in_control_boundary,omitempty" jsonschema:"Is this protect surface in the control boundary (are you responsible for its security)"`
	InZeroTrustFocus        *bool             `json:"in_zero_trust_focus,omitempty" jsonschema:"Is this protect surface in focus, should the security be measured and reported"`
	Relevance               *int              `json:"relevance,omitempty" jsonschema:"How important (0=not-100=very) is this protect service"`
	Confidentiality         *int              `json:"confidentiality,omitempty" jsonschema:"Score normally between 1-5"`
	Integrity               *int              `json:"integrity,omitempty" jsonschema:"Score normally between 1-5"`
	Availability            *int              `json:"availability,omitempty" jsonschema:"Score normally between 1-5"`
	DataTags                []string          `json:"data_tags,omitempty" jsonschema:"What data holds this protect surface, e.g. PII, PCI"`
	ComplianceTags          []string          `json:"compliance_tags,omitempty" jsonschema:"What compliance frameworks should be respected, e.g. GDPR, PCI-DSS, SOX"`
	CustomerLabels          map[string]string `json:"customer_labels,omitempty" jsonschema:"Customer specific Key=Value pairs for easy grouping and searching"`
}

// ListProtectSurfacesParams holds optional filters for listing protect surfaces.
// All fields are optional; with none supplied, every protect surface is returned.
// When multiple fields are supplied, all conditions must match (AND).
type ListProtectSurfacesParams struct {
	InZeroTrustFocus  *bool             `json:"in_zero_trust_focus,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces whose in_zero_trust_focus flag matches this value."`
	InControlBoundary *bool             `json:"in_control_boundary,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces whose in_control_boundary flag matches this value."`
	RelevanceMin      *int              `json:"relevance_min,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces with relevance >= this value (0-100)."`
	RelevanceMax      *int              `json:"relevance_max,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces with relevance <= this value (0-100)."`
	NameContains      string            `json:"name_contains,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces whose name contains this substring (case-insensitive)."`
	DataTags          []string          `json:"data_tags,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces whose data_tags contain ALL of these values (e.g. ['PII','PCI'])."`
	ComplianceTags    []string          `json:"compliance_tags,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces whose compliance_tags contain ALL of these values (e.g. ['GDPR','SOX'])."`
	CustomerLabels    map[string]string `json:"customer_labels,omitempty" jsonschema:"Optional filter. If set, only return protect surfaces whose customer_labels contain every supplied key=value pair."`
}

// LocationParams for location operations (create, update, and delete)
type LocationParams struct {
	ID            string   `json:"id,omitempty" jsonschema:"The ID of the location (required for update and delete operations)"`
	IDs           []string `json:"ids,omitempty" jsonschema:"Optional list of location IDs for bulk delete operations"`
	Name          string   `json:"name,omitempty" jsonschema:"The name of the location (required for create operations)"`
	UniquenessKey string   `json:"uniqueness_key,omitempty" jsonschema:"Key to prevent duplicates, especially in parallel processing"`
	Latitude      *float64 `json:"latitude,omitempty" jsonschema:"Latitude coordinate of the location"`
	Longitude     *float64 `json:"longitude,omitempty" jsonschema:"Longitude coordinate of the location"`
}

// StateParams for state operations (create, update, and delete)
type StateParams struct {
	ID               string   `json:"id,omitempty" jsonschema:"The ID of the state (required for update and delete operations)"`
	IDs              []string `json:"ids,omitempty" jsonschema:"Optional list of state IDs for bulk delete operations"`
	UniquenessKey    string   `json:"uniqueness_key,omitempty" jsonschema:"Key to prevent duplicates, especially in parallel processing"`
	Description      string   `json:"description,omitempty" jsonschema:"Description of the state - good practice since multiple states of same type can be attached to a Protect Surface"`
	ProtectSurface   string   `json:"protectsurface_id,omitempty" jsonschema:"Attached protect surface ID (required for create operations)"`
	ContentType      string   `json:"content_type,omitempty" jsonschema:"Content Type: ipv4, ipv6, azure_cloud, aws_cloud, gcp_cloud, container, hostname, user_identity (default: ipv4)"`
	Location         string   `json:"location_id,omitempty" jsonschema:"Attached location ID (required for create operations)"`
	ExistsOnAssetIDs []string `json:"exists_on_asset_ids,omitempty" jsonschema:"IDs of the managed assets providing this information"`
	Maintainer       string   `json:"maintainer,omitempty" jsonschema:"Either: portal_manual or api"`
	Content          []string `json:"content,omitempty" jsonschema:"The actual content e.g. ['10.10.10.1', '10.10.10.2']"`
}

// MeasureParams for measure operations (generally for listing available measures)
type MeasureParams struct {
	ID          string `json:"id,omitempty" jsonschema:"Exact measure ID to retrieve a single measure (e.g. 'ztmm-1.1')"`
	Category    string `json:"category,omitempty" jsonschema:"Exact match on measure group name, label or caption to filter by category"`
	SearchQuery string `json:"search_query,omitempty" jsonschema:"Case-insensitive search term matched against measure name, caption and explanation"`
}

// ProtectSurfaceMeasureParams for protect surface measure operations (list (only ID), update, remove (only ID))
type ProtectSurfaceMeasureParams struct {
	ProtectSurfaceID string `json:"protect_surface_id,omitempty" jsonschema:"The ID of the protect surface (required for all operations)"`
	MeasureName      string `json:"measure_name,omitempty" jsonschema:"The name/ID of the measure from the catalog (required for update, remove operations)"`

	// For filtering list operations
	AssignedOnly    *bool `json:"assigned_only,omitempty" jsonschema:"Filter to show only assigned measures"`
	ImplementedOnly *bool `json:"implemented_only,omitempty" jsonschema:"Filter to show only implemented measures"`

	// For assignment operations
	Assigned              *bool   `json:"assigned,omitempty" jsonschema:"Whether the measure is assigned to the protect surface"`
	AssignmentPersonEmail *string `json:"assignment_person_email,omitempty" jsonschema:"Email of person responsible for assignment. REQUIRED when assigned=true, otherwise optional. Use contacts to find email if needed."`

	// For implementation operations
	Implemented               *bool   `json:"implemented,omitempty" jsonschema:"Whether the measure is implemented"`
	ImplementationPersonEmail *string `json:"implementation_person_email,omitempty" jsonschema:"Email of person who validated implementation. REQUIRED when implemented=true, otherwise optional. Use contacts to find email if needed."`

	// For evidence operations
	Evidenced           *bool   `json:"evidenced,omitempty" jsonschema:"Whether the implementation has been evidenced as working"`
	EvidencePersonEmail *string `json:"evidence_person_email,omitempty" jsonschema:"Email of person who gathered evidence. REQUIRED when evidenced=true, otherwise optional. Use contacts to find email if needed."`

	// For risk acceptance operations
	RiskNoImplementationAccepted *bool   `json:"risk_no_implementation_accepted,omitempty" jsonschema:"Whether risk of no implementation validation is accepted"`
	RiskNoEvidenceAccepted       *bool   `json:"risk_no_evidence_accepted,omitempty" jsonschema:"Whether risk of no evidence is accepted"`
	RiskAcceptedComment          *string `json:"risk_accepted_comment,omitempty" jsonschema:"Comment explaining why risk is accepted. REQUIRED when any risk acceptance field is true, otherwise optional."`
	RiskAcceptancePersonEmail    *string `json:"risk_acceptance_person_email,omitempty" jsonschema:"Email of person who accepted risks. REQUIRED when any risk acceptance field is true, otherwise optional. Use contacts to find email if needed."`

	// Common fields
	UniquenessKey string `json:"uniqueness_key,omitempty" jsonschema:"Key to prevent duplicates, especially in parallel processing"`
}

// TransactionFlowParams for transaction flow operations between protect surfaces
type TransactionFlowParams struct {
	SourceProtectSurfaceID      string `json:"source_protect_surface_id,omitempty" jsonschema:"The ID of the source protect surface (required for create and delete operations)"`
	DestinationProtectSurfaceID string `json:"destination_protect_surface_id,omitempty" jsonschema:"The ID of the destination protect surface (required for create and delete operations)"`
	Allow                       *bool  `json:"allow,omitempty" jsonschema:"Whether the flow is allowed (true) or denied (false) - required for create operations"`
}

// ExternalFlowParams for external flow operations (to/from outside)
type ExternalFlowParams struct {
	ProtectSurfaceID string `json:"protect_surface_id,omitempty" jsonschema:"The ID of the protect surface (required for all operations)"`
	Direction        string `json:"direction,omitempty" jsonschema:"Flow direction: 'inbound' (from outside) or 'outbound' (to outside) - required for all operations"`
	Allow            *bool  `json:"allow,omitempty" jsonschema:"Whether the external flow is allowed (true) or denied (false) - required for create operations"`
}

// FlowListParams for listing flows of a protect surface
type FlowListParams struct {
	ProtectSurfaceID string `json:"protect_surface_id,omitempty" jsonschema:"The ID of the protect surface to list flows for (required)"`
}

// CaseParams for case/ticket operations (create, update, close, get)
type CaseParams struct {
	// For create and get operations
	ID string `json:"id,omitempty" jsonschema:"Optional customer-supplied identifier used for create. Required for update, close, and get operations."`

	// For create operation
	Subject             string `json:"subject,omitempty" jsonschema:"The subject/title of the case (required for create)"`
	Note                string `json:"note,omitempty" jsonschema:"Initial note or description for the case (required for create)"`
	Priority            *int   `json:"priority,omitempty" jsonschema:"Case priority from 1 to 4, where 1 is the highest priority (required for create, optional for priority update)"`
	CaseType            string `json:"case_type,omitempty" jsonschema:"The type of case: 'securityincident', 'incident', 'change', 'standardchange', 'inforequest' (required for create)"`
	PrimaryContactEmail *string `json:"primary_contact_email,omitempty" jsonschema:"Email address of the primary contact - must match a user in the system (required for create, optional for contact update)"`

	// For update operations
	NewSubject             string `json:"new_subject,omitempty" jsonschema:"New subject/title when updating the case subject"`
	NewPriority            *int   `json:"new_priority,omitempty" jsonschema:"New priority (1-4) when updating case priority"`
	NewPrimaryContactEmail string `json:"new_primary_contact_email,omitempty" jsonschema:"New primary contact email when updating the primary contact"`

	// For add note operation
	AdditionalNote string `json:"additional_note,omitempty" jsonschema:"Additional note to add to an existing case"`
}
