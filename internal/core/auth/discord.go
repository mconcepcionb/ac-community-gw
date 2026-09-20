package auth

import (
	"context"
	"time"
)

// DiscordUser is the subset of the Discord user object the gateway needs.
//
// The Discord ID is the stable identity key; username and global name are
// mutable profile data and are never used as identity.
type DiscordUser struct {
	ID         string
	Username   string
	GlobalName string
	Avatar     string
}

// DiscordToken is an OAuth2 token response.
//
// It is secret: it must never be logged, persisted, published in an event,
// written to the audit log or placed in a cookie.
type DiscordToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

// DiscordProvider is the transport contract for Discord OAuth2.
//
// It is implemented by internal/adapters/discord and consumed by the
// identity-discord plugin. Keeping the contract in core lets the plugin depend
// on an interface instead of importing the adapter, which the architecture
// tests forbid.
type DiscordProvider interface {
	// AuthorizeURL builds the authorization-code redirect URL.
	AuthorizeURL(state, codeChallenge string) string
	// Exchange trades an authorization code for an access token.
	Exchange(ctx context.Context, code, codeVerifier string) (DiscordToken, error)
	// CurrentUser returns the authenticated user's profile.
	CurrentUser(ctx context.Context, accessToken string) (DiscordUser, error)
	// GuildMemberRoleIDs returns the user's role ids in a guild.
	GuildMemberRoleIDs(ctx context.Context, accessToken, guildID string) ([]string, error)
}
