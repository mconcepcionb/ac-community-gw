// Package persistence contains readiness contracts shared by the core and its
// PostgreSQL adapter.
package persistence

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// defaultCheckTimeout bounds each individual readiness check.
const defaultCheckTimeout = 5 * time.Second

// Check is a single readiness probe (for example PostgreSQL connectivity).
type Check interface {
	Name() string
	Check(ctx context.Context) error
}

// Result is the outcome of one readiness check.
type Result struct {
	Name string
	Err  error
}

// Registry aggregates readiness checks.
type Registry struct {
	mu      sync.RWMutex
	checks  []Check
	timeout time.Duration
}

// NewRegistry creates an empty readiness registry.
func NewRegistry() *Registry {
	return &Registry{timeout: defaultCheckTimeout}
}

// SetTimeout overrides the per-check timeout. A non-positive value disables it.
func (r *Registry) SetTimeout(timeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timeout = timeout
}

// Add appends a readiness check.
func (r *Registry) Add(check Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks = append(r.checks, check)
}

// Evaluate runs every check, bounding each one with the configured timeout.
func (r *Registry) Evaluate(ctx context.Context) []Result {
	r.mu.RLock()
	checks := append([]Check(nil), r.checks...)
	timeout := r.timeout
	r.mu.RUnlock()

	results := make([]Result, 0, len(checks))
	for _, check := range checks {
		checkCtx := ctx
		cancel := func() {}
		if timeout > 0 {
			checkCtx, cancel = context.WithTimeout(ctx, timeout)
		}
		err := check.Check(checkCtx)
		cancel()
		results = append(results, Result{Name: check.Name(), Err: err})
	}
	return results
}

// Check runs all readiness checks and joins their errors.
func (r *Registry) Check(ctx context.Context) error {
	var errs []error
	for _, result := range r.Evaluate(ctx) {
		if result.Err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", result.Name, result.Err))
		}
	}
	return errors.Join(errs...)
}

// Names returns the registered check names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.checks))
	for _, check := range r.checks {
		names = append(names, check.Name())
	}
	return names
}
