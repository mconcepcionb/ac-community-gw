package identitydiscord

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
)

// provision upserts the community user for an authenticated Discord identity.
func (p *Plugin) provision(r *http.Request, user auth.DiscordUser) (uuid.UUID, bool, error) {
	return p.repo.Provision(r.Context(), user)
}
