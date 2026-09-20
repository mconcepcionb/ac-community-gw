// Package azerothdb defines read-only access to AzerothCore's databases.
//
// AzerothCore keeps accounts in its login database (MySQL/MariaDB), which is
// separate from the SOAP command interface. The SOAP interface has no command
// to list every account, so listing requires reading that database. The
// interface here is transport-agnostic and implemented by an adapter.
package azerothdb

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrUnavailable is returned when the login database cannot be reached.
var ErrUnavailable = errors.New("azerothdb: login database unavailable")

// ErrAccountNotFound is returned when a username does not exist.
var ErrAccountNotFound = errors.New("azerothdb: account not found")

// Account is a read-only view of an AzerothCore login account.
type Account struct {
	ID        int64
	Username  string
	Email     string
	GMLevel   int
	Expansion int
	Online    bool
	Banned    bool
	BanReason string
	LastIP    string
	LastLogin *time.Time
}

// Query filters and paginates an account listing.
type Query struct {
	// Filter matches usernames with a case-insensitive prefix/substring.
	Filter string
	Limit  int
	Offset int
}

// AccountReader reads AzerothCore login accounts.
type AccountReader interface {
	ListAccounts(ctx context.Context, query Query) ([]Account, error)
	// FindAccountByUsername returns a single account or ErrAccountNotFound.
	FindAccountByUsername(ctx context.Context, username string) (Account, error)
}

// Unavailable is an AccountReader used when a login database is configured but
// could not be reached. It lets callers tell "unavailable" apart from "not
// configured" (a nil reader).
type Unavailable struct {
	Err error
}

// ListAccounts implements AccountReader.
func (u Unavailable) ListAccounts(context.Context, Query) ([]Account, error) {
	if u.Err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return nil, ErrUnavailable
}

// FindAccountByUsername implements AccountReader.
func (u Unavailable) FindAccountByUsername(context.Context, string) (Account, error) {
	if u.Err != nil {
		return Account{}, fmt.Errorf("%w: %v", ErrUnavailable, u.Err)
	}
	return Account{}, ErrUnavailable
}
