package permissions

import (
	"sort"
	"sync"
)

// Authorizer resolves role -> permission grants.
//
// It is deliberately an in-memory structure for the first iteration. A
// PostgreSQL-backed implementation can replace it later without changing the
// authorization middleware contract.
type Authorizer struct {
	mu        sync.RWMutex
	rolePerms map[Role]map[Permission]struct{}
}

// NewAuthorizer creates an empty authorizer.
func NewAuthorizer() *Authorizer {
	return &Authorizer{rolePerms: make(map[Role]map[Permission]struct{})}
}

// Grant assigns permissions to a role, creating the role if needed.
func (a *Authorizer) Grant(role Role, permissions ...Permission) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.rolePerms[role] == nil {
		a.rolePerms[role] = make(map[Permission]struct{})
	}
	for _, p := range permissions {
		a.rolePerms[role][p] = struct{}{}
	}
}

// Replace atomically replaces every role -> permission grant. It is used to
// reload grants from the database without restarting the gateway.
func (a *Authorizer) Replace(rolePerms map[Role][]Permission) {
	next := make(map[Role]map[Permission]struct{}, len(rolePerms))
	for role, perms := range rolePerms {
		set := make(map[Permission]struct{}, len(perms))
		for _, p := range perms {
			set[p] = struct{}{}
		}
		next[role] = set
	}
	a.mu.Lock()
	a.rolePerms = next
	a.mu.Unlock()
}

// Can reports whether any of the given roles grants p.
func (a *Authorizer) Can(roles []Role, p Permission) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, role := range roles {
		if _, ok := a.rolePerms[role][p]; ok {
			return true
		}
	}
	return false
}

// Permissions returns the union of permissions granted to roles, sorted.
func (a *Authorizer) Permissions(roles []Role) []Permission {
	a.mu.RLock()
	defer a.mu.RUnlock()
	seen := make(map[Permission]struct{})
	for _, role := range roles {
		for p := range a.rolePerms[role] {
			seen[p] = struct{}{}
		}
	}
	out := make([]Permission, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// RoleNames returns the roles known to the authorizer, sorted.
func (a *Authorizer) RoleNames() []Role {
	a.mu.RLock()
	defer a.mu.RUnlock()
	roles := make([]Role, 0, len(a.rolePerms))
	for role := range a.rolePerms {
		roles = append(roles, role)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
	return roles
}
