package reports

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/domain"
)

var (
	errUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"reports_unavailable", "reports are unavailable")
	errInvalidReport = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_report", "target, category and message are required")
	errReportNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"report_not_found", "report not found")
)

const maxMessageLen = 1000

// SubmitReportRequest is the body of POST /api/v1/reports.
type SubmitReportRequest struct {
	Target   string `json:"target"`
	Category string `json:"category"`
	Message  string `json:"message"`
} // @name SubmitReportRequest

// Report is the JSON representation of a community report.
type Report struct {
	ID         string `json:"id"`
	ReporterID string `json:"reporter_id"`
	Target     string `json:"target"`
	Category   string `json:"category"`
	Message    string `json:"message"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
} // @name CommunityReport

// ReportsResponse is the body of the report listing endpoints.
type ReportsResponse struct {
	Reports []Report `json:"reports"`
} // @name CommunityReportsResponse

// handleCreate handles POST /api/v1/reports.
//
//	@Summary		Submit a report
//	@Description	Reports another player. Requires the gw.report.create permission.
//	@Tags			reports
//	@ID				reports.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	SubmitReportRequest	true	"report"
//	@Success		201	{object}	Report
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/reports [post]
func (p *Plugin) handleCreate(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	var req SubmitReportRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.Target = strings.TrimSpace(req.Target)
	req.Category = strings.TrimSpace(req.Category)
	req.Message = strings.TrimSpace(req.Message)
	if req.Target == "" || req.Category == "" || req.Message == "" || len(req.Message) > maxMessageLen {
		httpapi.WriteError(w, r, errInvalidReport)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	report, err := p.store.Create(r.Context(), principal.UserID, req.Target, req.Category, req.Message)
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         "report.create",
		Permission:     string(PermissionCreate),
		TargetType:     "character",
		TargetID:       req.Target,
		Result:         audit.ResultSuccess,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
	httpapi.WriteJSON(w, http.StatusCreated, reportDTO(report))
}

// handleMine handles GET /api/v1/reports/mine.
//
//	@Summary		My reports
//	@Description	Lists the authenticated user's own reports. Requires the gw.report.create permission.
//	@Tags			reports
//	@ID				reports.mine
//	@Produce		json
//	@Param			limit	query	int	false	"page size"	default(50)
//	@Param			offset	query	int	false	"page offset"	default(0)
//	@Success		200	{object}	ReportsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/reports/mine [get]
func (p *Plugin) handleMine(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	reports, err := p.store.ByReporter(r.Context(), principal.UserID,
		intParam(r, "limit", 50), intParam(r, "offset", 0))
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	writeReports(w, r, reports)
}

// handleList handles GET /api/v1/admin/reports.
//
//	@Summary		List reports
//	@Description	Lists every report, newest first, optionally filtered by status. Requires the gw.report.read permission.
//	@Tags			reports
//	@ID				reports.list
//	@Produce		json
//	@Param			status	query	string	false	"report status (open, closed)"
//	@Param			limit	query	int		false	"page size"	default(50)
//	@Param			offset	query	int		false	"page offset"	default(0)
//	@Success		200	{object}	ReportsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/reports [get]
func (p *Plugin) handleList(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	reports, err := p.store.List(r.Context(), status,
		intParam(r, "limit", 50), intParam(r, "offset", 0))
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	writeReports(w, r, reports)
}

// handleClose handles POST /api/v1/admin/reports/{id}/close.
//
//	@Summary		Close a report
//	@Description	Marks a report closed. Requires the gw.report.read permission.
//	@Tags			reports
//	@ID				reports.close
//	@Produce		json
//	@Param			id	path	string	true	"report id"
//	@Success		200	{object}	Report
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/reports/{id}/close [post]
func (p *Plugin) handleClose(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	report, err := p.store.Close(r.Context(), id)
	if errors.Is(err, domain.ErrReportNotFound) {
		httpapi.WriteError(w, r, errReportNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, reportDTO(report))
}

func writeReports(w http.ResponseWriter, r *http.Request, reports []domain.Report) {
	items := make([]Report, 0, len(reports))
	for _, report := range reports {
		items = append(items, reportDTO(report))
	}
	httpapi.WriteJSON(w, http.StatusOK, ReportsResponse{Reports: items})
}

func reportDTO(report domain.Report) Report {
	return Report{
		ID:         report.ID.String(),
		ReporterID: report.ReporterID.String(),
		Target:     report.Target,
		Category:   report.Category,
		Message:    report.Message,
		Status:     report.Status,
		CreatedAt:  report.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func intParam(r *http.Request, name string, fallback int) int {
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
