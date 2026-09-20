// Package delivery defines the cross-plugin capability to deliver in-game mail,
// items and money to a character owned by an account.
package delivery

import (
	"context"
	"errors"
)

var (
	// ErrNotOwner is returned when the character does not belong to the account.
	ErrNotOwner = errors.New("delivery: character does not belong to the account")
	// ErrCharacterNotFound is returned when the character does not exist.
	ErrCharacterNotFound = errors.New("delivery: character not found")
	// ErrInvalidRequest is returned for a malformed delivery request.
	ErrInvalidRequest = errors.New("delivery: invalid request")
)

// Item is a single item stack to deliver.
type Item struct {
	ID    int
	Count int
}

// Request describes an in-game delivery.
type Request struct {
	// AccountID is the login account that must own the character.
	AccountID int64
	Character string
	Subject   string
	Body      string
	Items     []Item
	Money     int64
}

// Service delivers mail, items and money. It is implemented by azeroth-character
// and consumed by other plugins (for example azeroth-store).
type Service interface {
	Deliver(ctx context.Context, req Request) (output string, err error)
}
