// Package auth provides authentication primitives: principals, server-side
// sessions and cookie handling.
//
// Authentication establishes who the user is (Discord OAuth2). Authorization
// (what the user may do) lives in the permissions package and is enforced by
// the HTTP middleware.
package auth

import (
	"context"

	"github.com/google/uuid"
)

// Principal is an authenticated community user.
type Principal struct {
	UserID    uuid.UUID
	DiscordID string
	Roles     []string
	// Permissions carries explicit scopes for service credentials (API keys).
	// Interactive users derive permissions from their roles instead.
	Permissions []string
}

// APIKeyAuthenticator resolves a raw API key to a principal.
type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, rawKey string) (Principal, bool)
}

type principalKey struct{}

// WithPrincipal stores the principal in the context.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFromContext extracts the principal, if any.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}
