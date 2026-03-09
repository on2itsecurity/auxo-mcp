package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/on2itsecurity/auxo-mcp/server/config"
	"github.com/on2itsecurity/go-auxo/v2"
)

// Manager handles AUXO client creation
type Manager struct {
	config *config.Config
}

// NewManager creates a new client manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{config: cfg}
}

// CreateClient creates a new AUXO client using the Zero Trust token
func (m *Manager) CreateClient(ctx context.Context) (*auxo.Client, error) {
	overrides := OverridesFromContext(ctx)
	token := m.config.ZTToken
	if overrides.ZTToken != "" {
		token = overrides.ZTToken
	}
	return m.createClientWithToken(ctx, token, "AUXO_ZT_TOKEN")
}

// CreateCaseClient creates a new AUXO client using the ticket token
func (m *Manager) CreateCaseClient(ctx context.Context) (*auxo.Client, error) {
	overrides := OverridesFromContext(ctx)
	token := m.config.TicketToken
	if overrides.TicketToken != "" {
		token = overrides.TicketToken
	}
	return m.createClientWithToken(ctx, token, "AUXO_TICKET_TOKEN")
}

// createClientWithToken creates an AUXO client with the given token and applies shared overrides
func (m *Manager) createClientWithToken(ctx context.Context, token string, tokenName string) (*auxo.Client, error) {
	overrides := OverridesFromContext(ctx)

	apiURL := m.config.APIUrl
	if overrides.APIURL != "" {
		apiURL = overrides.APIURL
	}
	if apiURL == "" {
		apiURL = config.DefaultAPIURL
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("missing %s: supply the token in the request URL or configure %s before starting the server", tokenName, tokenName)
	}

	debug := m.config.Debug
	if overrides.Debug != nil {
		debug = *overrides.Debug
	}

	client, err := auxo.NewClient(apiURL, token, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise AUXO client: %w", err)
	}

	return client, nil
}
