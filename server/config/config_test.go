package config

import "testing"

// clearEnv sets every AUXO_* variable Load reads to "" for the duration of the
// test, so values from the developer's shell do not leak into assertions.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"AUXO_API_URL", "AUXO_ZT_TOKEN", "AUXO_TICKET_TOKEN", "AUXO_SERVER_MODE",
		"AUXO_SERVER_PORT", "AUXO_ENABLE_ZERO_TRUST", "AUXO_ENABLE_TICKETS", "AUXO_DEBUG",
	} {
		t.Setenv(k, "")
	}
}

// TestEmptyEnvBehavesLikeUnset covers the Claude plugin case, where every
// variable is always passed and a blank token arrives as an empty string.
func TestEmptyEnvBehavesLikeUnset(t *testing.T) {
	blanks := map[string]string{
		"empty":       "",
		"whitespace":  "   ",
		"placeholder": "${user_config.zt_token}",
	}
	for name, blank := range blanks {
		t.Run(name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("AUXO_ZT_TOKEN", blank)
			t.Setenv("AUXO_TICKET_TOKEN", blank)
			t.Setenv("AUXO_API_URL", blank)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load returned error for blank values: %v", err)
			}
			if cfg.ZTToken != "" || cfg.TicketToken != "" {
				t.Errorf("tokens = %q, %q; want both empty", cfg.ZTToken, cfg.TicketToken)
			}
			if cfg.EnableZeroTrust || cfg.EnableTickets {
				t.Errorf("domains enabled (zt=%t, tickets=%t); want both disabled", cfg.EnableZeroTrust, cfg.EnableTickets)
			}
			if cfg.APIUrl != DefaultAPIURL {
				t.Errorf("APIUrl = %q; want %q", cfg.APIUrl, DefaultAPIURL)
			}
			if cfg.ServerMode != "STDIO" {
				t.Errorf("ServerMode = %q; want STDIO", cfg.ServerMode)
			}
		})
	}
}

// TestOneTokenEnablesOneDomain checks that a blank token for one domain does
// not interfere with a real token for the other.
func TestOneTokenEnablesOneDomain(t *testing.T) {
	clearEnv(t)
	t.Setenv("AUXO_ZT_TOKEN", " your-zt-token ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.EnableZeroTrust || cfg.EnableTickets {
		t.Errorf("domains (zt=%t, tickets=%t); want zt only", cfg.EnableZeroTrust, cfg.EnableTickets)
	}
	if cfg.ZTToken != "your-zt-token" {
		t.Errorf("ZTToken = %q; want trimmed value", cfg.ZTToken)
	}
}

// TestCustomAPIURL checks that a configured endpoint still overrides the default.
func TestCustomAPIURL(t *testing.T) {
	clearEnv(t)
	t.Setenv("AUXO_API_URL", "api.example.net")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIUrl != "api.example.net" {
		t.Errorf("APIUrl = %q; want api.example.net", cfg.APIUrl)
	}
}
