package identitydiscord

import "github.com/google/uuid"

const (
	// EventUserAuthenticated is emitted after a successful Discord login.
	EventUserAuthenticated = "identity.user_authenticated"
	// EventDiscordRolesChanged is emitted when a user's Discord roles change.
	EventDiscordRolesChanged = "identity.discord_roles_changed"
	// EventUserProvisioned is emitted when a community user is first created.
	EventUserProvisioned = "identity.user_provisioned"
)

// UserAuthenticated announces that a user completed authentication.
type UserAuthenticated struct {
	UserID    uuid.UUID
	DiscordID string
}

// EventName implements events.Event.
func (UserAuthenticated) EventName() string { return EventUserAuthenticated }

// DiscordRolesChanged announces that a user's Discord roles were synchronised.
type DiscordRolesChanged struct {
	UserID    uuid.UUID
	DiscordID string
	Roles     []string
}

// EventName implements events.Event.
func (DiscordRolesChanged) EventName() string { return EventDiscordRolesChanged }

// UserProvisioned announces that a community user was created.
type UserProvisioned struct {
	UserID    uuid.UUID
	DiscordID string
}

// EventName implements events.Event.
func (UserProvisioned) EventName() string { return EventUserProvisioned }
