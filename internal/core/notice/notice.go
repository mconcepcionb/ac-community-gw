// Package notice defines the cross-plugin capability to send a short in-game
// text notice (subject + body) to a character.
//
// AzerothCore mail requires an enclosure, so implementations attach a
// configured notice item; the capability itself is value-free.
package notice

import (
	"context"
	"errors"
)

var (
	// ErrNotOwner is returned when the character does not belong to the account.
	ErrNotOwner = errors.New("notice: character does not belong to the account")
	// ErrCharacterNotFound is returned when the character does not exist.
	ErrCharacterNotFound = errors.New("notice: character not found")
	// ErrInvalidRequest is returned for a malformed notice request.
	ErrInvalidRequest = errors.New("notice: invalid request")
	// ErrNotConfigured is returned when notices are disabled.
	ErrNotConfigured = errors.New("notice: not configured")
)

// Request describes an in-game notice.
type Request struct {
	// AccountID is the login account that must own the character.
	AccountID int64
	Character string
	Subject   string
	Body      string
}

// Service sends a text notice to a character owned by an account.
type Service interface {
	Send(ctx context.Context, req Request) (output string, err error)
}
