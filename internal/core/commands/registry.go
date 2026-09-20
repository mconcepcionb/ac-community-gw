// Package commands implements the application command registry.
//
// Application commands describe gateway capabilities (for example
// "account.change-password"), never AzerothCore CLI/SOAP syntax. The plugin
// that owns a capability registers the handler; any other plugin may execute
// it through this registry without importing the implementation.
package commands

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	// ErrNotFound is returned when executing an unregistered command.
	ErrNotFound = errors.New("commands: not found")
	// ErrAlreadyRegistered is returned when registering a duplicate name.
	ErrAlreadyRegistered = errors.New("commands: already registered")
	// ErrEmptyName is returned for an empty command name.
	ErrEmptyName = errors.New("commands: name must not be empty")
	// ErrInvalidPayload is returned when the payload type does not match.
	ErrInvalidPayload = errors.New("commands: invalid payload")
	// ErrInvalidResult is returned when a result type does not match.
	ErrInvalidResult = errors.New("commands: invalid result")
)

// Handler executes a command with an untyped payload and returns an untyped
// result.
type Handler func(ctx context.Context, payload any) (any, error)

// Registry is a concurrency-safe set of named command handlers.
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

// Register adds a handler under name.
func (r *Registry) Register(name string, handler Handler) error {
	if name == "" {
		return ErrEmptyName
	}
	if handler == nil {
		return errors.New("commands: handler must not be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, name)
	}
	r.handlers[name] = handler
	return nil
}

// Execute runs the handler registered for name.
func (r *Registry) Execute(ctx context.Context, name string, payload any) (any, error) {
	r.mu.RLock()
	handler, ok := r.handlers[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return handler(ctx, payload)
}

// Has reports whether a command is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[name]
	return ok
}

// Names returns the registered command names in sorted order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterTyped registers a strongly typed handler under name.
func RegisterTyped[P any, R any](r *Registry, name string, fn func(ctx context.Context, payload P) (R, error)) error {
	return r.Register(name, func(ctx context.Context, payload any) (any, error) {
		typed, ok := payload.(P)
		if !ok {
			var zero P
			return nil, fmt.Errorf("%w: expected %T, got %T", ErrInvalidPayload, zero, payload)
		}
		return fn(ctx, typed)
	})
}

// ExecuteTyped runs a command and asserts the typed result.
func ExecuteTyped[P any, R any](ctx context.Context, r *Registry, name string, payload P) (R, error) {
	var zero R
	result, err := r.Execute(ctx, name, payload)
	if err != nil {
		return zero, err
	}
	typed, ok := result.(R)
	if !ok {
		return zero, fmt.Errorf("%w: expected %T, got %T", ErrInvalidResult, zero, result)
	}
	return typed, nil
}
