package azerothaccount

import (
	"net/http"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

// record writes an audit entry for an account command outcome. It records the
// target username only; passwords and emails are never stored.
func (p *Plugin) record(r *http.Request, action, targetID string, err error) {
	principal, _ := auth.PrincipalFromContext(r.Context())
	result := audit.ResultSuccess
	if err != nil {
		result = audit.ResultFailure
	}
	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         action,
		Permission:     string(PermissionAccountManage),
		TargetType:     "account",
		TargetID:       targetID,
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}
