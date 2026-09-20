// Package services implements the synchronous service/capability registry.
//
// A plugin publishes a capability (a plain Go value implementing an interface)
// under a stable name. Consumers resolve it and assert it to an interface they
// declare in their own package. This keeps Go's structural typing as the
// contract mechanism and avoids cross-plugin imports.
package services

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	// ErrNotFound is returned when resolving an unpublished capability.
	ErrNotFound = errors.New("services: not found")
	// ErrAlreadyPublished is returned when publishing a duplicate name.
	ErrAlreadyPublished = errors.New("services: already published")
	// ErrEmptyName is returned for an empty capability name.
	ErrEmptyName = errors.New("services: name must not be empty")
)

// Registry is a concurrency-safe set of named capabilities.
type Registry struct {
	mu     sync.RWMutex
	values map[string]any
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{values: make(map[string]any)}
}

// Publish registers a capability under name.
func (r *Registry) Publish(name string, service any) error {
	if name == "" {
		return ErrEmptyName
	}
	if service == nil {
		return errors.New("services: capability must not be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.values[name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyPublished, name)
	}
	r.values[name] = service
	return nil
}

// Resolve returns the capability published under name.
func (r *Registry) Resolve(name string) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.values[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}

// Has reports whether a capability is published.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.values[name]
	return ok
}

// Names returns the published capability names in sorted order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.values))
	for name := range r.values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Provide publishes a typed capability.
func Provide[T any](r *Registry, name string, service T) error {
	return r.Publish(name, service)
}

// Consume resolves a capability and asserts it to T.
func Consume[T any](r *Registry, name string) (T, error) {
	var zero T
	value, err := r.Resolve(name)
	if err != nil {
		return zero, err
	}
	typed, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("services: capability %q has type %T, not %T", name, value, zero)
	}
	return typed, nil
}
