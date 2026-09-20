package plugins

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrEmptyName is returned for a plugin without a name.
	ErrEmptyName = errors.New("plugins: name must not be empty")
	// ErrDuplicateName is returned for two plugins with the same name.
	ErrDuplicateName = errors.New("plugins: duplicate name")
)

// Manager registers plugins in a deterministic order.
type Manager struct {
	plugins []Plugin
}

// NewManager creates an empty manager.
func NewManager() *Manager {
	return &Manager{}
}

// Add appends a plugin.
func (m *Manager) Add(plugin Plugin) {
	m.plugins = append(m.plugins, plugin)
}

// RegisterAll registers every plugin against the registry.
func (m *Manager) RegisterAll(ctx context.Context, reg *Registry) error {
	seen := make(map[string]struct{}, len(m.plugins))
	for _, plugin := range m.plugins {
		name := plugin.Name()
		if name == "" {
			return ErrEmptyName
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateName, name)
		}
		seen[name] = struct{}{}
		if err := plugin.Register(ctx, reg); err != nil {
			return fmt.Errorf("plugins: register %s: %w", name, err)
		}
	}
	return nil
}

// Names returns the plugin names in registration order.
func (m *Manager) Names() []string {
	names := make([]string, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		names = append(names, plugin.Name())
	}
	return names
}
