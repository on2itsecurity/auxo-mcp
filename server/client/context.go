package client

import "context"

// Overrides allows per-request configuration of AUXO credentials
type Overrides struct {
	APIURL          string
	ZTToken         string
	TicketToken     string
	EnableZeroTrust *bool
	EnableTickets   *bool
	Debug           *bool
}

type contextKey struct{}

var overridesKey = contextKey{}

func ContextWithOverrides(ctx context.Context, overrides Overrides) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if overrides.APIURL == "" && overrides.ZTToken == "" && overrides.TicketToken == "" && overrides.EnableZeroTrust == nil && overrides.EnableTickets == nil && overrides.Debug == nil {
		return ctx
	}
	return context.WithValue(ctx, overridesKey, overrides)
}

func OverridesFromContext(ctx context.Context) Overrides {
	if ctx == nil {
		return Overrides{}
	}
	if overrides, ok := ctx.Value(overridesKey).(Overrides); ok {
		return overrides
	}
	return Overrides{}
}
