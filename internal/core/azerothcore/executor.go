// Package azerothcore defines the transport contract for executing
// AzerothCore administrative commands.
//
// The interface knows nothing about SOAP or CLI syntax: it only transports an
// already-built command string. Command ownership (which plugin knows which
// syntax) lives in the plugins.
package azerothcore

import (
	"context"
	"errors"
)

// CommandExecutor transports an AzerothCore administrative command.
//
// Implementations must not know domain semantics; they only exchange the
// command with AzerothCore and translate transport errors.
type CommandExecutor interface {
	Execute(ctx context.Context, command string) (string, error)
}

// ErrNotConfigured is returned when no AzerothCore executor is configured.
var ErrNotConfigured = errors.New("azerothcore: command executor not configured")

// Unavailable is a CommandExecutor used when no SOAP endpoint is configured.
type Unavailable struct{}

// Execute implements CommandExecutor.
func (Unavailable) Execute(context.Context, string) (string, error) {
	return "", ErrNotConfigured
}
