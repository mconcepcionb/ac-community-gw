// Package domain holds the azeroth-account module's shared types.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrLinkNotFound is returned when a user has no linked account.
	ErrLinkNotFound = errors.New("azerothaccount: account link not found")
	// ErrAccountAlreadyLinked is returned when an account links another user.
	ErrAccountAlreadyLinked = errors.New("azerothaccount: account already linked")
)

// Link is a community user's link to an AzerothCore account.
type Link struct {
	UserID          uuid.UUID
	AccountUsername string
	AccountID       *int64
	LinkedAt        time.Time
	UpdatedAt       time.Time
}
