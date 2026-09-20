package azerothadmin

import (
	"net/http"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
)

// record writes an audit entry for an administrative command outcome.
func (p *Plugin) record(r *http.Request, action string, permission permissions.Permission, targetType, targetID string, err error) {
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
		Permission:     string(permission),
		TargetType:     targetType,
		TargetID:       targetID,
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}
