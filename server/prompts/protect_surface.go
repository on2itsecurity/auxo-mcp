package prompts

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
)

// ProtectSurfacePrompts handles protect surface related prompts
type ProtectSurfacePrompts struct {
	clientManager *client.Manager
}

// NewProtectSurfacePrompts creates a new instance of ProtectSurfacePrompts
func NewProtectSurfacePrompts(clientManager *client.Manager) *ProtectSurfacePrompts {
	return &ProtectSurfacePrompts{
		clientManager: clientManager,
	}
}

// CreateWithLocationAndState generates a prompt for creating a protect surface with location and state
func (p *ProtectSurfacePrompts) CreateWithLocationAndState(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Extract arguments from the request
	args := make(map[string]string)
	if req.Params.Arguments != nil {
		args = req.Params.Arguments
	}

	// Get optional arguments for templating
	protectSurfaceName := getStringArgFromStringMap(args, "protect_surface_name", "MyProtectSurface")
	locationName := getStringArgFromStringMap(args, "location_name", "MyLocation")
	contentType := getStringArgFromStringMap(args, "content_type", "ipv4")
	content := getStringArgFromStringMap(args, "content", "10.0.0.1,10.0.0.2")

	// Build the prompt content
	promptContent := fmt.Sprintf(`I'll help you create a protect surface along with its location and state. Here's a step-by-step guide:

## Creating a Protect Surface with Location and State

### Step 1: Create the Location
First, we need to create a location where the protect surface will be situated.

**Example Location Creation:**
- Name: %s
- Coordinates: You can optionally specify latitude and longitude
- Uniqueness Key: Use a unique identifier to prevent duplicates

### Step 2: Create the Protect Surface
Next, we'll create the protect surface itself.

**Example Protect Surface Creation:**
- Name: %s
- Description: Brief description of what this protect surface protects
- Confidentiality: Score 1-5 (how sensitive is the data)
- Integrity: Score 1-5 (how critical is data accuracy)
- Availability: Score 1-5 (how critical is uptime)
- Relevance: Score 0-100 (importance level)
- In Control Boundary: true/false (are you responsible for its security)
- In Zero Trust Focus: true/false (should security be measured and reported)
- Data Tags: Array of data types like ["PII", "PCI", "Financial"]
- Compliance Tags: Array of compliance frameworks like ["GDPR", "PCI-DSS", "SOX"]

### Step 3: Create the State
Finally, we'll create a state that contains the actual content/assets for this protect surface.

**Example State Creation:**
- Content Type: %s (ipv4, ipv6, azure_cloud, aws_cloud, gcp_cloud, container, hostname, user_identity)
- Content: [%s] (actual IP addresses, hostnames, or other identifiers)
- Description: Description of what this state represents
- Location ID: Links to the location created in Step 1
- Protect Surface ID: Links to the protect surface created in Step 2

### Workflow Commands:
1. Use **createLocation** tool to create the location
2. Use **createProtectSurface** tool to create the protect surface (note the returned ID)
3. Use **createState** tool to create the state, referencing both location and protect surface IDs

### Important Notes:
- Save the IDs returned from each creation step as they're needed for linking
- The location and protect surface must exist before creating the state
- Use meaningful names and descriptions for better organization
- Consider the security scores carefully as they affect risk assessments

Would you like me to help you create these resources with specific values?`, locationName, protectSurfaceName, contentType, content)

	// Create the prompt result
	result := &mcp.GetPromptResult{
		Description: "A comprehensive guide for creating a protect surface with its associated location and state",
		Messages: []*mcp.PromptMessage{
			{
				Role: mcp.Role("assistant"),
				Content: &mcp.TextContent{
					Text: promptContent,
				},
			},
		},
	}

	return result, nil
}

// Helper function to get string argument with default value
func getStringArgFromStringMap(args map[string]string, key, defaultValue string) string {
	if val, ok := args[key]; ok {
		return val
	}
	return defaultValue
}
