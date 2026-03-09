package client

import (
	"fmt"
	"strings"
)

// FriendlyError normalizes AUXO API errors for consumption by LLM clients.
func FriendlyError(err error) error {
	if err == nil {
		return nil
	}

	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "401"), strings.Contains(msg, "unauthorized"):
		return fmt.Errorf("AUXO rejected the provided token (HTTP 401 Unauthorized). Provide a valid token via ?token=... or configure AUXO_ZT_TOKEN before starting the server")
	case strings.Contains(msg, "403"), strings.Contains(msg, "forbidden"):
		return fmt.Errorf("AUXO denied access for the provided token (HTTP 403 Forbidden). Ensure the token has the necessary permissions or request a new one")
	case strings.Contains(msg, "invalid token"), strings.Contains(msg, "token is malformed"):
		return fmt.Errorf("The provided AUXO token appears to be malformed. Double-check the value passed via ?token=... or AUXO_ZT_TOKEN")
	default:
		return err
	}
}
