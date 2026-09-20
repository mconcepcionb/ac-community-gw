package identitydiscord

import (
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

const (
	metricLoginStarted   = "identity_login_started_total"
	metricLoginCompleted = "identity_login_completed_total"
	metricLoginFailed    = "identity_login_failed_total"
	metricSessionCreated = "identity_session_created_total"
	metricRolesChanged   = "identity_roles_changed_total"
)

func (p *Plugin) inc(name string) {
	if p.metrics == nil {
		return
	}
	p.metrics.Inc(name)
}

// loginFailed counts a failed login attempt and writes the error envelope.
func (p *Plugin) loginFailed(w http.ResponseWriter, r *http.Request, err error) {
	p.inc(metricLoginFailed)
	httpapi.WriteError(w, r, err)
}
