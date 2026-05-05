package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/config"
	"github.com/on2itsecurity/auxo-mcp/server/prompts"
	"github.com/on2itsecurity/auxo-mcp/server/resources"
	"github.com/on2itsecurity/auxo-mcp/server/tools"
)

// queryOverrides holds parsed query parameter overrides for HTTP mode
type queryOverrides struct {
	ztToken         string
	ticketToken     string
	apiURL          string
	enableZeroTrust *bool
	enableTickets   *bool
	debug           *bool
}

// parseQueryOverrides extracts override values from URL query parameters
func parseQueryOverrides(query url.Values, cfg *config.Config) queryOverrides {
	var o queryOverrides

	o.ztToken = strings.TrimSpace(query.Get("zt_token"))
	if o.ztToken == "" {
		// Support legacy 'token' parameter for backward compatibility
		o.ztToken = strings.TrimSpace(query.Get("token"))
	}
	o.ticketToken = strings.TrimSpace(query.Get("ticket_token"))

	o.apiURL = strings.TrimSpace(query.Get("api"))
	if o.apiURL == "" {
		o.apiURL = strings.TrimSpace(query.Get("url"))
	}

	if ztParam := query.Get("enable_zero_trust"); ztParam != "" {
		ztBool := config.ParseBool(ztParam, cfg.EnableZeroTrust)
		o.enableZeroTrust = &ztBool
	}

	if ticketParam := query.Get("enable_tickets"); ticketParam != "" {
		ticketBool := config.ParseBool(ticketParam, cfg.EnableTickets)
		o.enableTickets = &ticketBool
	}

	if debugParam := query.Get("debug"); debugParam != "" {
		debugBool := config.ParseBool(debugParam, cfg.Debug)
		o.debug = &debugBool
	}

	return o
}

// hasOverrides returns true if any override is set
func (o queryOverrides) hasOverrides() bool {
	return o.ztToken != "" || o.ticketToken != "" || o.apiURL != "" || o.enableZeroTrust != nil || o.enableTickets != nil || o.debug != nil
}

// applyTo creates an effective config by applying overrides to the base config
func (o queryOverrides) applyTo(base *config.Config) *config.Config {
	effectiveCfg := &config.Config{
		APIUrl:          base.APIUrl,
		ZTToken:         base.ZTToken,
		TicketToken:     base.TicketToken,
		ServerMode:      base.ServerMode,
		ServerPort:      base.ServerPort,
		EnableZeroTrust: base.EnableZeroTrust,
		EnableTickets:   base.EnableTickets,
		Debug:           base.Debug,
	}

	if o.apiURL != "" {
		effectiveCfg.APIUrl = o.apiURL
	}
	if o.ztToken != "" {
		effectiveCfg.ZTToken = o.ztToken
	}
	if o.ticketToken != "" {
		effectiveCfg.TicketToken = o.ticketToken
	}
	if o.enableZeroTrust != nil {
		effectiveCfg.EnableZeroTrust = *o.enableZeroTrust
	}
	if o.enableTickets != nil {
		effectiveCfg.EnableTickets = *o.enableTickets
	}
	if o.debug != nil {
		effectiveCfg.Debug = *o.debug
	}

	return effectiveCfg
}

// autoDetectDomains enables domains when a query provides a token but no explicit enable flag
func (o queryOverrides) autoDetectDomains(cfg *config.Config) {
	if o.enableZeroTrust == nil && o.ztToken != "" {
		cfg.EnableZeroTrust = true
	}
	if o.enableTickets == nil && o.ticketToken != "" {
		cfg.EnableTickets = true
	}
}

// toClientOverrides converts to client.Overrides for context propagation
func (o queryOverrides) toClientOverrides() client.Overrides {
	return client.Overrides{
		ZTToken:         o.ztToken,
		TicketToken:     o.ticketToken,
		APIURL:          o.apiURL,
		EnableZeroTrust: o.enableZeroTrust,
		EnableTickets:   o.enableTickets,
		Debug:           o.debug,
	}
}

// Build-time variables (injected via ldflags)
var (
	version   = "dev"     // Version tag
	commit    = "unknown" // Git commit hash
	buildTime = "unknown" // Build timestamp
)

func main() {
	// Define command line flags
	modeFlag := flag.String("mode", "", "Server mode: STDIO or HTTP (overrides env vars)")
	portFlag := flag.Int("port", 0, "Port for HTTP mode (overrides env vars)")
	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version if requested
	if *versionFlag {
		fmt.Printf("AUXO MCP Server %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", buildTime)
		return
	}

	// Show help if requested
	if *helpFlag {
		fmt.Printf("AUXO MCP Server %s\n", version)
		fmt.Println("Usage: auxo-mcp-server [flags]")
		fmt.Println("\nConfiguration Priority (highest to lowest):")
		fmt.Println("  1. Command line flags (-mode, -port)")
		fmt.Println("  2. Environment variables (AUXO_*)")
		fmt.Println("\nEnvironment Variables:")
		fmt.Println("  AUXO_API_URL           - AUXO API endpoint (default: api.on2it.net)")
		fmt.Println("  AUXO_ZT_TOKEN          - Zero Trust API token")
		fmt.Println("  AUXO_TICKET_TOKEN      - Ticket API token")
		fmt.Println("  AUXO_SERVER_MODE       - Server mode: STDIO or HTTP (default: STDIO)")
		fmt.Println("  AUXO_SERVER_PORT       - Port for HTTP mode (default: 8080)")
		fmt.Println("  AUXO_ENABLE_ZERO_TRUST - Enable Zero Trust domain (auto-detected from token, override with true/false)")
		fmt.Println("  AUXO_ENABLE_TICKETS    - Enable Tickets domain (auto-detected from token, override with true/false)")
		fmt.Println("  AUXO_DEBUG             - Enable debug logging for API calls (true/false, default: false)")
		fmt.Println("\nHTTP Mode Query Parameters (for clients that can't use environment variables):")
		fmt.Println("  ?zt_token=...          - Override Zero Trust token (legacy: ?token=...)")
		fmt.Println("  ?ticket_token=...      - Override Ticket token")
		fmt.Println("  ?api=...               - Override API endpoint (alias: ?url=...)")
		fmt.Println("  ?enable_zero_trust=... - Override Zero Trust domain enable flag")
		fmt.Println("  ?enable_tickets=...    - Override Tickets domain enable flag")
		fmt.Println("  ?debug=...             - Override debug mode (true/false)")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  auxo-mcp-server                        # Use environment variables")
		fmt.Println("  auxo-mcp-server -mode STDIO            # Force STDIO mode")
		fmt.Println("  auxo-mcp-server -mode HTTP             # Force HTTP mode with configured port")
		fmt.Println("  auxo-mcp-server -mode HTTP -port 9090  # Force HTTP mode on port 9090")
		fmt.Println("\nMCP Configuration Example (.vscode/mcp.json):")
		fmt.Println(`  {
    "servers": {
      "auxo-mcp-server": {
        "type": "stdio",
        "command": "path/to/auxo-mcp-server",
        "env": {
          "AUXO_API_URL": "api.on2it.net",
          "AUXO_ZT_TOKEN": "your-token-here",
          "AUXO_TICKET_TOKEN": "your-ticket-token"
        }
      }
    }
  }`)
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Override configuration with command line flags if provided
	if *modeFlag != "" {
		if *modeFlag != "STDIO" && *modeFlag != "HTTP" {
			log.Fatalf("Invalid mode '%s'. Must be either STDIO or HTTP", *modeFlag)
		}
		cfg.ServerMode = *modeFlag
		log.Printf("Mode overridden via command line: %s", *modeFlag)
	}

	if *portFlag != 0 {
		cfg.ServerPort = *portFlag
		log.Printf("Port overridden via command line: %d", *portFlag)
	}

	// Helper function to create and configure a server instance based on effective configuration
	createServer := func(effectiveCfg *config.Config) *mcp.Server {
		// Create a new client manager for this effective configuration
		effectiveClientManager := client.NewManager(effectiveCfg)

		// Create server with explicit capabilities
		serverOptions := &mcp.ServerOptions{
			// Explicitly disable logging capabilities to prevent client-side setLevel errors
			// This tells VS Code not to attempt logging/setLevel during initialization
		}

		server := mcp.NewServer(&mcp.Implementation{
			Name:    "AUXO MCP Server",
			Version: version},
			serverOptions)

		// Initialize and register Zero Trust tools (only if enabled)
		if effectiveCfg.EnableZeroTrust {
			protectSurfaceTools := tools.NewProtectSurfaceTools(effectiveClientManager)
			locationTools := tools.NewLocationTools(effectiveClientManager)
			stateTools := tools.NewStateTools(effectiveClientManager)
			contactTools := tools.NewContactTools(effectiveClientManager)
			assetTools := tools.NewAssetTools(effectiveClientManager)
			measureTools := tools.NewMeasureTools(effectiveClientManager)
			protectSurfaceMeasureTools := tools.NewProtectSurfaceMeasureTools(effectiveClientManager)
			transactionFlowTools := tools.NewTransactionFlowTools(effectiveClientManager)
			protectSurfacePrompts := prompts.NewProtectSurfacePrompts(effectiveClientManager)

			// Initialize resource handlers
			resourceHandlers := resources.NewHandlers(effectiveClientManager)

			// Register tools - Protect Surfaces
			mcp.AddTool(server, &mcp.Tool{Name: "createProtectSurface", Description: "Create a new protect surface"}, protectSurfaceTools.Create)
			mcp.AddTool(server, &mcp.Tool{Name: "listProtectSurfaces", Description: "Get All Protect Surfaces, returns ID, Name and Relevance"}, protectSurfaceTools.List)
			mcp.AddTool(server, &mcp.Tool{Name: "getProtectSurface", Description: "Get full details of a protect surface by its ID"}, protectSurfaceTools.Get)
			mcp.AddTool(server, &mcp.Tool{Name: "updateProtectSurface", Description: "Update an existing protect surface"}, protectSurfaceTools.Update)
			mcp.AddTool(server, &mcp.Tool{Name: "deleteProtectSurface", Description: "Delete one or more protect surfaces"}, protectSurfaceTools.Delete)

			// Register tools - Locations
			mcp.AddTool(server, &mcp.Tool{Name: "createLocation", Description: "Create a new location"}, locationTools.Create)
			mcp.AddTool(server, &mcp.Tool{Name: "listLocations", Description: "Get All Locations"}, locationTools.List)
			mcp.AddTool(server, &mcp.Tool{Name: "updateLocation", Description: "Update an existing location"}, locationTools.Update)
			mcp.AddTool(server, &mcp.Tool{Name: "deleteLocation", Description: "Delete one or more locations"}, locationTools.Delete)

			// Register tools - States
			mcp.AddTool(server, &mcp.Tool{Name: "createState", Description: "Create a new state"}, stateTools.Create)
			mcp.AddTool(server, &mcp.Tool{Name: "listStates", Description: "Get All States"}, stateTools.List)
			mcp.AddTool(server, &mcp.Tool{Name: "updateState", Description: "Update an existing state"}, stateTools.Update)
			mcp.AddTool(server, &mcp.Tool{Name: "deleteState", Description: "Delete one or more states"}, stateTools.Delete)

			// Register tools - Contacts & Assets
			mcp.AddTool(server, &mcp.Tool{Name: "listContacts", Description: "Get All Contacts"}, contactTools.List)
			mcp.AddTool(server, &mcp.Tool{Name: "listAssets", Description: "Get All Assets"}, assetTools.List)

			// Register tools - Measures
			mcp.AddTool(server, &mcp.Tool{Name: "listMeasures", Description: "List/search measures from the AUXO catalog. Filters: 'id' for exact measure ID lookup (e.g. 'ztmm-1.1'), 'category' for exact group/category match, 'search_query' for case-insensitive text search across name, caption and explanation. Without filters returns all measures. There is no framework filter; framework mappings are included in each measure's data."}, measureTools.List)

			// Register tools - Protect Surface Measures
			mcp.AddTool(server, &mcp.Tool{Name: "listProtectSurfaceMeasures", Description: "List measures assigned to a protect surface"}, protectSurfaceMeasureTools.List)
			mcp.AddTool(server, &mcp.Tool{Name: "updateProtectSurfaceMeasure", Description: "Add or Update the implementation status of a measure on a protect surface"}, protectSurfaceMeasureTools.Update)
			mcp.AddTool(server, &mcp.Tool{Name: "removeMeasureFromProtectSurface", Description: "Remove a measure assignment from a protect surface"}, protectSurfaceMeasureTools.Remove)

			// Register tools - Transaction Flows
			mcp.AddTool(server, &mcp.Tool{Name: "createTransactionFlow", Description: "Create a flow between two protect surfaces with mutual consensus"}, transactionFlowTools.CreateFlow)
			mcp.AddTool(server, &mcp.Tool{Name: "createExternalFlow", Description: "Create a flow to/from outside the organization for a protect surface"}, transactionFlowTools.CreateExternalFlow)
			mcp.AddTool(server, &mcp.Tool{Name: "listTransactionFlows", Description: "List all flows for a specific protect surface"}, transactionFlowTools.ListFlows)
			mcp.AddTool(server, &mcp.Tool{Name: "deleteTransactionFlow", Description: "Delete a flow between two protect surfaces with mutual consensus"}, transactionFlowTools.DeleteFlow)
			mcp.AddTool(server, &mcp.Tool{Name: "deleteExternalFlow", Description: "Delete an external flow (to/from outside) for a protect surface"}, transactionFlowTools.DeleteExternalFlow)

			// Register prompts
			server.AddPrompt(&mcp.Prompt{
				Name:        "create-protect-surface-with-location-and-state",
				Title:       "Create Protect Surface with Location and State",
				Description: "A comprehensive guide for creating a protect surface along with its associated location and state",
				Arguments: []*mcp.PromptArgument{
					{
						Name:        "protect_surface_name",
						Description: "Name for the protect surface",
						Required:    false,
					},
					{
						Name:        "location_name",
						Description: "Name for the location",
						Required:    false,
					},
					{
						Name:        "content_type",
						Description: "Type of content (ipv4, ipv6, azure_cloud, aws_cloud, gcp_cloud, container, hostname, user_identity)",
						Required:    false,
					},
					{
						Name:        "content",
						Description: "Comma-separated content values",
						Required:    false,
					},
				},
			}, protectSurfacePrompts.CreateWithLocationAndState)

			// Register resources
			server.AddResource(resourceHandlers.ProtectSurfacesResource(), resourceHandlers.ProtectSurfaces)
			server.AddResource(resourceHandlers.ContactsResource(), resourceHandlers.Contacts)
			server.AddResource(resourceHandlers.LocationsResource(), resourceHandlers.Locations)
			server.AddResource(resourceHandlers.StatesResource(), resourceHandlers.States)
			server.AddResource(resourceHandlers.AssetsResource(), resourceHandlers.Assets)
			server.AddResource(resourceHandlers.TransactionFlowsResource(), resourceHandlers.TransactionFlows)
			server.AddResource(resourceHandlers.MeasuresResource(), resourceHandlers.Measures)

			log.Println("Zero Trust domain enabled")
		}

		// Initialize and register Tickets tools (only if enabled)
		if effectiveCfg.EnableTickets {
			caseTools := tools.NewCaseTools(effectiveClientManager)

			// Register case tools
			mcp.AddTool(server, &mcp.Tool{Name: "createCase", Description: "Create a new support case/ticket in the system. Required fields: id, subject, note, priority (1-4 where 1=highest), primary_contact_email, and case_type. Valid case types: 'securityincident' (security incident requiring immediate attention), 'incident' (service disruption or issue), 'change' (planned change requiring approval), 'standardchange' (pre-approved routine change), 'inforequest' (information or assistance request). It is wise to provide a friendly id for tracking, it is not the case-number that will be autogenerated."}, caseTools.Create)
			mcp.AddTool(server, &mcp.Tool{Name: "getCases", Description: "Get all cases/tickets"}, caseTools.GetAll)
			mcp.AddTool(server, &mcp.Tool{Name: "getCase", Description: "Get details of a case/ticket by its ID"}, caseTools.Get)
			mcp.AddTool(server, &mcp.Tool{Name: "updateCasePriority", Description: "Update the priority (1-4) of an existing case, where 1 is highest priority"}, caseTools.UpdatePriority)
			mcp.AddTool(server, &mcp.Tool{Name: "updateCasePrimaryContact", Description: "Update the primary contact email for an existing case"}, caseTools.UpdatePrimaryContact)
			mcp.AddTool(server, &mcp.Tool{Name: "updateCaseSubject", Description: "Update the subject/title of an existing case"}, caseTools.UpdateSubject)
			mcp.AddTool(server, &mcp.Tool{Name: "escalateCase", Description: "Escalate an existing case to higher priority attention"}, caseTools.Escalate)
			mcp.AddTool(server, &mcp.Tool{Name: "deescalateCase", Description: "De-escalate an existing case back to normal handling"}, caseTools.Deescalate)
			mcp.AddTool(server, &mcp.Tool{Name: "addNoteToCase", Description: "Add a note/comment to an existing case"}, caseTools.AddNote)
			mcp.AddTool(server, &mcp.Tool{Name: "closeCase", Description: "Request to close a case. The case will be reviewed by an engineer and closed if no further work is needed."}, caseTools.Close)

			log.Println("Tickets domain enabled")
		}

		return server
	}

	// Run the server
	switch cfg.ServerMode {
	case "STDIO":
		log.Println("Starting MCP server in STDIO mode; awaiting client connection on stdin/stdout")
		// Reduce log verbosity in STDIO mode to minimize client warnings
		server := createServer(cfg)
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	case "HTTP":
		log.Printf("Starting MCP server in HTTP mode on port %d", cfg.ServerPort)

		// Create SSE handler that creates a new server instance per request
		sseHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
			o := parseQueryOverrides(request.URL.Query(), cfg)
			effectiveCfg := o.applyTo(cfg)
			o.autoDetectDomains(effectiveCfg)

			if o.hasOverrides() {
				log.Printf("Creating server with request overrides (zt_token: %t, ticket_token: %t, api: %t, enable_zero_trust: %v, enable_tickets: %v, debug: %v)",
					o.ztToken != "", o.ticketToken != "", o.apiURL != "", effectiveCfg.EnableZeroTrust, effectiveCfg.EnableTickets, effectiveCfg.Debug)
			}

			return createServer(effectiveCfg)
		}, nil)

		// Setup HTTP server
		http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
			o := parseQueryOverrides(r.URL.Query(), cfg)
			effectiveCfg := o.applyTo(cfg)
			o.autoDetectDomains(effectiveCfg)

			// Validate required tokens for enabled domains
			if effectiveCfg.EnableZeroTrust && effectiveCfg.ZTToken == "" {
				http.Error(w, "Zero Trust domain is enabled but missing token. Provide ?zt_token=YOUR_ZT_TOKEN in the request URL, configure AUXO_ZT_TOKEN before starting the server, or disable Zero Trust with ?enable_zero_trust=false", http.StatusUnauthorized)
				return
			}

			if effectiveCfg.EnableTickets && effectiveCfg.TicketToken == "" {
				http.Error(w, "Tickets domain is enabled but missing token. Provide ?ticket_token=YOUR_TICKET_TOKEN in the request URL, configure AUXO_TICKET_TOKEN before starting the server, or disable Tickets with ?enable_tickets=false", http.StatusUnauthorized)
				return
			}

			ctx := client.ContextWithOverrides(r.Context(), o.toClientOverrides())
			sseHandler.ServeHTTP(w, r.WithContext(ctx))
		})

		// Start HTTP server
		addr := fmt.Sprintf(":%d", cfg.ServerPort)
		log.Printf("MCP server listening on %s/sse", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Unsupported server mode, either use SSE or STDIO: %s", cfg.ServerMode)
	}
}
