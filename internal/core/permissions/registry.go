// Package permissions provides the permission registry and RBAC checks.
//
// The core knows how to register permissions, grant them to roles and answer
// authorization questions. It never knows the concrete permissions owned by
// plugins.
package permissions

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Permission is a stable, namespaced authorization identifier.
type Permission string

// Role is an internal role granted to a user.
type Role string

var (
	// ErrEmptyName is returned for an empty permission name.
	ErrEmptyName = errors.New("permissions: name must not be empty")
	// ErrAlreadyRegistered is returned when registering a duplicate permission.
	ErrAlreadyRegistered = errors.New("permissions: already registered")
	// ErrNamespaceMismatch is returned when a permission name does not start
	// with its declared namespace.
	ErrNamespaceMismatch = errors.New("permissions: name does not match namespace")
)

// Definition describes a permission owned by a plugin.
type Definition struct {
	Name        Permission
	Description string
	Owner       string
	// Namespace is the required name prefix: "gw" for gateway-generic
	// capabilities or a registered game id (for example "azeroth"). The name
	// must start with Namespace + ".". See ADR 0014.
	Namespace string
}

// Registry is a concurrency-safe catalogue of permission definitions.
type Registry struct {
	mu   sync.RWMutex
	defs map[Permission]Definition
}

// NewRegistry creates an empty permission registry.
func NewRegistry() *Registry {
	return &Registry{defs: make(map[Permission]Definition)}
}

// Register adds a permission definition.
func (r *Registry) Register(def Definition) error {
	if def.Name == "" {
		return ErrEmptyName
	}
	if def.Namespace == "" || !strings.HasPrefix(string(def.Name), def.Namespace+".") {
		return fmt.Errorf("%w: %s is not in namespace %q", ErrNamespaceMismatch, def.Name, def.Namespace)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.defs[def.Name]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, def.Name)
	}
	r.defs[def.Name] = def
	return nil
}

// Has reports whether a permission is registered.
func (r *Registry) Has(p Permission) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.defs[p]
	return ok
}

// Definitions returns all definitions sorted by name.
func (r *Registry) Definitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]Definition, 0, len(r.defs))
	for _, def := range r.defs {
		defs = append(defs, def)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })
	return defs
}
