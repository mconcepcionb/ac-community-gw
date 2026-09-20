package azerothadmin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
)

var errAuditUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
	"audit_unavailable", "the audit log is unavailable")

// AuditEntry is one audit log row in the viewer.
type AuditEntry struct {
	OccurredAt     string `json:"occurred_at"`
	ActorID        string `json:"actor_id,omitempty"`
	ActorDiscordID string `json:"actor_discord_id,omitempty"`
	Action         string `json:"action"`
	Permission     string `json:"permission,omitempty"`
	TargetType     string `json:"target_type,omitempty"`
	TargetID       string `json:"target_id,omitempty"`
	Result         string `json:"result"`
	RequestID      string `json:"request_id,omitempty"`
} // @name AdminAuditEntry

// AuditResponse is the body of GET /api/v1/admin/audit.
type AuditResponse struct {
	Entries []AuditEntry `json:"entries"`
} // @name AdminAuditResponse

// handleAuditLog handles GET /api/v1/admin/audit.
//
//	@Summary		Read the audit log
//	@Description	Lists audit entries, newest first, filterable by actor, target, action and time range. Requires the audit.read permission.
//	@Tags			azeroth-admin
//	@ID				azeroth.admin.audit.list
//	@Produce		json
//	@Param			actor	query	string	false	"actor community user id"
//	@Param			target	query	string	false	"target id substring"
//	@Param			action	query	string	false	"action substring"
//	@Param			since	query	string	false	"RFC3339 lower bound"
//	@Param			until	query	string	false	"RFC3339 upper bound"
//	@Param			limit	query	int		false	"page size"	default(50)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	AuditResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/audit [get]
func (p *Plugin) handleAuditLog(w http.ResponseWriter, r *http.Request) {
	if p.auditReader == nil {
		httpapi.WriteError(w, r, errAuditUnavailable)
		return
	}
	query := r.URL.Query()
	filter := audit.ListFilter{
		ActorID: query.Get("actor"),
		Target:  query.Get("target"),
		Action:  query.Get("action"),
	}
	if since, err := time.Parse(time.RFC3339, query.Get("since")); err == nil {
		filter.Since = since
	}
	if until, err := time.Parse(time.RFC3339, query.Get("until")); err == nil {
		filter.Until = until
	}
	entries, err := p.auditReader.List(r.Context(), filter,
		intQuery(r, "limit", 50), intQuery(r, "offset", 0))
	if err != nil {
		httpapi.WriteError(w, r, errAuditUnavailable)
		return
	}
	items := make([]AuditEntry, 0, len(entries))
	for _, entry := range entries {
		items = append(items, auditDTO(entry))
	}
	httpapi.WriteJSON(w, http.StatusOK, AuditResponse{Entries: items})
}

func auditDTO(entry audit.Entry) AuditEntry {
	dto := AuditEntry{
		ActorDiscordID: entry.ActorDiscordID,
		Action:         entry.Action,
		Permission:     entry.Permission,
		TargetType:     entry.TargetType,
		TargetID:       entry.TargetID,
		Result:         string(entry.Result),
		RequestID:      entry.RequestID,
	}
	if entry.ActorID.String() != "00000000-0000-0000-0000-000000000000" {
		dto.ActorID = entry.ActorID.String()
	}
	if !entry.Timestamp.IsZero() {
		dto.OccurredAt = entry.Timestamp.UTC().Format(time.RFC3339)
	}
	return dto
}

func intQuery(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}
