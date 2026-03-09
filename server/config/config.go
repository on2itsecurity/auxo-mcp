package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// DefaultAPIURL is used when no API endpoint is configured
const DefaultAPIURL = "api.on2it.net"

// Config holds AUXO API settings and server configuration
type Config struct {
	APIUrl          string
	ZTToken         string
	TicketToken     string // For future use
	ServerMode      string // STDIO or HTTP
	ServerPort      int    // Port for HTTP mode
	EnableZeroTrust bool   // Enable Zero Trust domain tools/prompts/resources
	EnableTickets   bool   // Enable Tickets domain tools/prompts/resources
	Debug           bool   // Enable debug logging for API requests/responses
}

// Load configuration from environment variables only
func Load() (*Config, error) {
	config := &Config{
		// Set defaults
		APIUrl:     DefaultAPIURL,
		ServerMode: "STDIO",
		ServerPort: 8080,
		Debug:      false, // Default to disabled
	}

	// Load from environment variables
	if envAPIUrl := os.Getenv("AUXO_API_URL"); envAPIUrl != "" {
		config.APIUrl = envAPIUrl
	}
	if envZTToken := os.Getenv("AUXO_ZT_TOKEN"); envZTToken != "" {
		config.ZTToken = envZTToken
	}
	if envTicketToken := os.Getenv("AUXO_TICKET_TOKEN"); envTicketToken != "" {
		config.TicketToken = envTicketToken
	}
	if envServerMode := os.Getenv("AUXO_SERVER_MODE"); envServerMode != "" {
		config.ServerMode = envServerMode
	}
	if envServerPort := os.Getenv("AUXO_SERVER_PORT"); envServerPort != "" {
		if port, err := strconv.Atoi(envServerPort); err == nil && port > 0 {
			config.ServerPort = port
		}
	}

	// Auto-detect domain enablement based on token presence
	config.EnableZeroTrust = config.ZTToken != ""
	config.EnableTickets = config.TicketToken != ""

	// Allow explicit override via environment variables
	if envEnableZT := os.Getenv("AUXO_ENABLE_ZERO_TRUST"); envEnableZT != "" {
		config.EnableZeroTrust = ParseBool(envEnableZT, config.EnableZeroTrust)
	}
	if envEnableTickets := os.Getenv("AUXO_ENABLE_TICKETS"); envEnableTickets != "" {
		config.EnableTickets = ParseBool(envEnableTickets, config.EnableTickets)
	}

	// Load debug flag
	if envDebug := os.Getenv("AUXO_DEBUG"); envDebug != "" {
		config.Debug = ParseBool(envDebug, false)
	}

	// Ensure defaults remain if environment overrides cleared them
	if config.APIUrl == "" {
		config.APIUrl = DefaultAPIURL
	}

	// Validate server mode
	if config.ServerMode != "STDIO" && config.ServerMode != "HTTP" {
		return nil, fmt.Errorf("server mode must be either STDIO or HTTP, got: %s", config.ServerMode)
	}

	// Validate that required tokens are provided for explicitly enabled domains
	if config.EnableZeroTrust && config.ZTToken == "" {
		return nil, fmt.Errorf("AUXO_ZT_TOKEN is required when Zero Trust domain is enabled")
	}
	if config.EnableTickets && config.TicketToken == "" {
		return nil, fmt.Errorf("AUXO_TICKET_TOKEN is required when Tickets domain is enabled")
	}

	// Warn if no domains are enabled
	if !config.EnableZeroTrust && !config.EnableTickets {
		log.Println("Warning: No domains enabled. Provide AUXO_ZT_TOKEN and/or AUXO_TICKET_TOKEN to enable features.")
	}

	return config, nil
}

// ParseBool parses a boolean string with a default value
func ParseBool(s string, defaultVal bool) bool {
	switch s {
	case "true", "True", "TRUE", "1", "yes", "Yes", "YES":
		return true
	case "false", "False", "FALSE", "0", "no", "No", "NO":
		return false
	default:
		return defaultVal
	}
}
