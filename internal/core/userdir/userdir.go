// Package userdir defines the cross-plugin capability to resolve and search
// community users owned by identity-discord.
package userdir

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User is a community user with its Discord profile.
type User struct {
	ID          uuid.UUID
	DiscordID   string
	Username    string
	GlobalName  string
	DisplayName string
	CreatedAt   time.Time
}

// Directory resolves and searches community users.
type Directory interface {
	// ResolveUser resolves a Discord id to a community user id.
	ResolveUser(ctx context.Context, discordID string) (userID uuid.UUID, found bool, err error)
	// ResolveUserByName returns the users whose Discord username, global name
	// or display name matches exactly (case-insensitive).
	ResolveUserByName(ctx context.Context, name string) ([]User, error)
	// ListUsers searches users by id, username, global name or display name.
	ListUsers(ctx context.Context, filter string, limit, offset int) ([]User, error)
}
