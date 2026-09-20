package identitydiscord

import (
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
)

// MeResponse is the authenticated principal returned by GET /api/v1/me.
type MeResponse struct {
	UserID      string   `json:"user_id"`
	DiscordID   string   `json:"discord_id"`
	Username    string   `json:"username"`
	GlobalName  string   `json:"global_name"`
	DisplayName string   `json:"display_name"`
	Avatar      string   `json:"avatar"`
	CreatedAt   string   `json:"created_at"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
} // @name MeResponse

// CallbackResponse is returned by the OAuth callback when no post-login
// redirect is configured.
type CallbackResponse struct {
	UserID    string   `json:"user_id"`
	DiscordID string   `json:"discord_id"`
	Roles     []string `json:"roles"`
} // @name CallbackResponse

// effectivePermissions resolves the union of permissions granted to roles.
// It always returns a non-nil slice so the JSON is an array, never null.
func (p *Plugin) effectivePermissions(roles []string) []string {
	if p.authorizer == nil {
		return []string{}
	}
	converted := make([]permissions.Role, 0, len(roles))
	for _, role := range roles {
		converted = append(converted, permissions.Role(role))
	}
	granted := p.authorizer.Permissions(converted)
	out := make([]string, 0, len(granted))
	for _, permission := range granted {
		out = append(out, string(permission))
	}
	return out
}
