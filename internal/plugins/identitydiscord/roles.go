package identitydiscord

import (
	"net/http"
	"sort"

	"github.com/google/uuid"
)

// syncRoles maps the user's Discord guild roles to internal roles and persists
// the result. It returns the effective roles and whether they changed.
//
// When no guild is configured it falls back to the last known roles. Only
// roles present in discord_role_mappings ever reach the principal.
func (p *Plugin) syncRoles(r *http.Request, accessToken string, userID uuid.UUID) ([]string, bool, error) {
	previous, err := p.repo.UserRoles(r.Context(), userID)
	if err != nil {
		return nil, false, err
	}
	previous = normalizeRoles(previous)

	if p.provider == nil || p.guildID == "" {
		return previous, false, nil
	}

	ids, err := p.provider.GuildMemberRoleIDs(r.Context(), accessToken, p.guildID)
	if err != nil {
		return nil, false, err
	}
	mapped, err := p.repo.MappedRoles(r.Context(), ids)
	if err != nil {
		return nil, false, err
	}
	// Static mappings from configuration are merged with the database ones.
	for _, id := range ids {
		if role, ok := p.roleMappings[id]; ok {
			mapped = append(mapped, role)
		}
	}
	mapped = normalizeRoles(mapped)

	changed := !sameRoles(previous, mapped)
	if changed {
		if err := p.repo.UpdateUserRoles(r.Context(), userID, mapped); err != nil {
			return nil, false, err
		}
	}
	return mapped, changed, nil
}

// normalizeRoles deduplicates and sorts roles so comparisons are stable.
func normalizeRoles(roles []string) []string {
	if len(roles) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(roles))
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	sort.Strings(out)
	return out
}

func sameRoles(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// rolesOrEmpty renders a nil role slice as an empty JSON array.
func rolesOrEmpty(roles []string) []string {
	if roles == nil {
		return []string{}
	}
	return roles
}
