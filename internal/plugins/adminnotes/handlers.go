package adminnotes

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
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/adminnotes/domain"
)

var (
	errUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"notes_unavailable", "annotations are unavailable")
	errInvalidAnnotation = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_annotation", "target_type, target_id and body are required")
	errAnnotationNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"annotation_not_found", "annotation not found")
)

const maxBodyLen = 2000

// CreateAnnotationRequest is the body of POST /api/v1/admin/annotations.
type CreateAnnotationRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Body       string `json:"body"`
} // @name AdminAnnotationRequest

// UpdateAnnotationRequest is the body of PATCH /api/v1/admin/annotations/{id}.
type UpdateAnnotationRequest struct {
	Body string `json:"body"`
} // @name AdminAnnotationUpdateRequest

// Annotation is the JSON representation of a staff annotation.
type Annotation struct {
	ID         string `json:"id"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	AuthorID   string `json:"author_id"`
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
} // @name AdminAnnotation

// AnnotationsResponse is the body of the annotation listing endpoint.
type AnnotationsResponse struct {
	Annotations []Annotation `json:"annotations"`
} // @name AdminAnnotationsResponse

// handleList handles GET /api/v1/admin/annotations.
//
//	@Summary		List annotations
//	@Description	Lists the staff annotations of one target, newest first. Requires gw.notes.read.
//	@Tags			admin-notes
//	@ID				admin.notes.list
//	@Produce		json
//	@Param			target_type	query	string	true	"target type (account, user, character)"
//	@Param			target_id	query	string	true	"target identifier"
//	@Param			limit		query	int		false	"page size"	default(50)
//	@Param			offset		query	int		false	"page offset"	default(0)
//	@Success		200	{object}	AnnotationsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/annotations [get]
func (p *Plugin) handleList(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	targetType := strings.TrimSpace(r.URL.Query().Get("target_type"))
	targetID := strings.TrimSpace(r.URL.Query().Get("target_id"))
	if !domain.ValidTargetType(targetType) || targetID == "" {
		httpapi.WriteError(w, r, errInvalidAnnotation)
		return
	}
	annotations, err := p.store.ByTarget(r.Context(), targetType, targetID,
		intParam(r, "limit", 50), intParam(r, "offset", 0))
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	items := make([]Annotation, 0, len(annotations))
	for _, annotation := range annotations {
		items = append(items, annotationDTO(annotation))
	}
	httpapi.WriteJSON(w, http.StatusOK, AnnotationsResponse{Annotations: items})
}

// handleCreate handles POST /api/v1/admin/annotations.
//
//	@Summary		Create an annotation
//	@Description	Attaches a staff note to a target. Requires gw.notes.write.
//	@Tags			admin-notes
//	@ID				admin.notes.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateAnnotationRequest	true	"annotation"
//	@Success		201	{object}	Annotation
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/annotations [post]
func (p *Plugin) handleCreate(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	var req CreateAnnotationRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.TargetType = strings.TrimSpace(req.TargetType)
	req.TargetID = strings.TrimSpace(req.TargetID)
	req.Body = strings.TrimSpace(req.Body)
	if !domain.ValidTargetType(req.TargetType) || req.TargetID == "" ||
		req.Body == "" || len(req.Body) > maxBodyLen {
		httpapi.WriteError(w, r, errInvalidAnnotation)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	annotation, err := p.store.Create(r.Context(), req.TargetType, req.TargetID, principal.UserID, req.Body)
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	p.record(r, principal, "admin.notes.create", annotation)
	httpapi.WriteJSON(w, http.StatusCreated, annotationDTO(annotation))
}

// handleUpdate handles PATCH /api/v1/admin/annotations/{id}.
//
//	@Summary		Update an annotation
//	@Description	Edits an annotation body. Authors may edit their own; gw.notes.manage may edit any. Requires gw.notes.write.
//	@Tags			admin-notes
//	@ID				admin.notes.update
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string					true	"annotation id"
//	@Param			request	body	UpdateAnnotationRequest	true	"annotation"
//	@Success		200	{object}	Annotation
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/annotations/{id} [patch]
func (p *Plugin) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	var req UpdateAnnotationRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" || len(req.Body) > maxBodyLen {
		httpapi.WriteError(w, r, errInvalidAnnotation)
		return
	}
	existing, err := p.store.Get(r.Context(), id)
	if errors.Is(err, domain.ErrAnnotationNotFound) {
		httpapi.WriteError(w, r, errAnnotationNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	if existing.AuthorID != principal.UserID && !p.canManage(principal.Roles) {
		httpapi.WriteError(w, r, httpapi.ErrForbidden)
		return
	}
	annotation, err := p.store.Update(r.Context(), id, req.Body)
	if errors.Is(err, domain.ErrAnnotationNotFound) {
		httpapi.WriteError(w, r, errAnnotationNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	p.record(r, principal, "admin.notes.update", annotation)
	httpapi.WriteJSON(w, http.StatusOK, annotationDTO(annotation))
}

// handleDelete handles DELETE /api/v1/admin/annotations/{id}.
//
//	@Summary		Delete an annotation
//	@Description	Deletes an annotation. Authors may delete their own; gw.notes.manage may delete any. Requires gw.notes.write.
//	@Tags			admin-notes
//	@ID				admin.notes.delete
//	@Produce		json
//	@Param			id	path	string	true	"annotation id"
//	@Success		204
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/annotations/{id} [delete]
func (p *Plugin) handleDelete(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	existing, err := p.store.Get(r.Context(), id)
	if errors.Is(err, domain.ErrAnnotationNotFound) {
		httpapi.WriteError(w, r, errAnnotationNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	principal, _ := auth.PrincipalFromContext(r.Context())
	if existing.AuthorID != principal.UserID && !p.canManage(principal.Roles) {
		httpapi.WriteError(w, r, httpapi.ErrForbidden)
		return
	}
	if err := p.store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrAnnotationNotFound) {
			httpapi.WriteError(w, r, errAnnotationNotFound)
			return
		}
		httpapi.WriteError(w, r, errUnavailable)
		return
	}
	p.record(r, principal, "admin.notes.delete", existing)
	w.WriteHeader(http.StatusNoContent)
}

func (p *Plugin) record(r *http.Request, principal auth.Principal, action string, annotation domain.Annotation) {
	_ = p.audit.Record(r.Context(), audit.Entry{
		Timestamp:      time.Now(),
		ActorID:        principal.UserID,
		ActorDiscordID: principal.DiscordID,
		Action:         action,
		Permission:     string(PermissionWrite),
		TargetType:     annotation.TargetType,
		TargetID:       annotation.TargetID,
		Result:         audit.ResultSuccess,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}

func annotationDTO(annotation domain.Annotation) Annotation {
	return Annotation{
		ID:         annotation.ID.String(),
		TargetType: annotation.TargetType,
		TargetID:   annotation.TargetID,
		AuthorID:   annotation.AuthorID.String(),
		Body:       annotation.Body,
		CreatedAt:  annotation.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  annotation.UpdatedAt.UTC().Format(time.RFC3339),
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
