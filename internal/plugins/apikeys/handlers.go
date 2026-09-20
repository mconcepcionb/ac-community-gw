package apikeys

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/apikeys/domain"
)

var (
	errKeysUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"apikeys_unavailable", "API keys are unavailable")
	errKeyNotFound = httpapi.NewAPIError(http.StatusNotFound,
		"api_key_not_found", "API key not found")
	errInvalidKey = httpapi.NewAPIError(http.StatusUnprocessableEntity,
		"invalid_api_key", "name and at least one permission are required")
)

// Permission describes a registered permission for the scope picker.
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
} // @name ApiPermission

// PermissionsResponse is the body of GET /api/v1/admin/permissions.
type PermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
} // @name ApiPermissionsResponse

// APIKey is the JSON representation of an issued key (no secret).
type APIKey struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	KeyPrefix   string   `json:"key_prefix"`
	Permissions []string `json:"permissions"`
	CreatedAt   string   `json:"created_at"`
	LastUsedAt  string   `json:"last_used_at,omitempty"`
} // @name ApiKey

// APIKeysResponse is the body of GET /api/v1/admin/api-keys.
type APIKeysResponse struct {
	Keys []APIKey `json:"keys"`
} // @name ApiKeysResponse

// CreateAPIKeyRequest is the body of POST /api/v1/admin/api-keys.
type CreateAPIKeyRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
} // @name CreateApiKeyRequest

// APIKeySecretResponse returns a key with its one-time secret.
type APIKeySecretResponse struct {
	Key    APIKey `json:"key"`
	Secret string `json:"secret"`
} // @name ApiKeySecretResponse

// handlePermissions handles GET /api/v1/admin/permissions.
//
//	@Summary		List registered permissions
//	@Description	Returns every permission registered by the plugins, for the API key scope picker. Requires the apikeys.manage permission.
//	@Tags			apikeys
//	@ID				apikeys.permissions.list
//	@Produce		json
//	@Success		200	{object}	PermissionsResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/permissions [get]
func (p *Plugin) handlePermissions(w http.ResponseWriter, r *http.Request) {
	if p.registry == nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	defs := p.registry.Permissions.Definitions()
	items := make([]Permission, 0, len(defs))
	for _, def := range defs {
		items = append(items, Permission{
			Name:        string(def.Name),
			Description: def.Description,
			Owner:       def.Owner,
		})
	}
	httpapi.WriteJSON(w, http.StatusOK, PermissionsResponse{Permissions: items})
}

// handleList handles GET /api/v1/admin/api-keys.
//
//	@Summary		List API keys
//	@Description	Lists every API key (without secrets). Requires the apikeys.manage permission.
//	@Tags			apikeys
//	@ID				apikeys.list
//	@Produce		json
//	@Success		200	{object}	APIKeysResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/api-keys [get]
func (p *Plugin) handleList(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	keys, err := p.store.List(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	items := make([]APIKey, 0, len(keys))
	for _, key := range keys {
		items = append(items, keyDTO(key))
	}
	httpapi.WriteJSON(w, http.StatusOK, APIKeysResponse{Keys: items})
}

// handleCreate handles POST /api/v1/admin/api-keys.
//
//	@Summary		Create an API key
//	@Description	Issues a scoped API key. The secret is returned once. Requires the apikeys.manage permission.
//	@Tags			apikeys
//	@ID				apikeys.create
//	@Accept			json
//	@Produce		json
//	@Param			request	body	CreateAPIKeyRequest	true	"name and scopes"
//	@Success		201	{object}	APIKeySecretResponse
//	@Failure		400	{object}	httpapi.ErrorResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/api-keys [post]
func (p *Plugin) handleCreate(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	var req CreateAPIKeyRequest
	if err := httpapi.DecodeJSON(r, &req); err != nil {
		httpapi.WriteError(w, r, err)
		return
	}
	name := strings.TrimSpace(req.Name)
	permissions := cleanPermissions(req.Permissions)
	if name == "" || len(permissions) == 0 {
		httpapi.WriteError(w, r, errInvalidKey)
		return
	}
	key, secret, err := p.store.Create(r.Context(), name, permissions)
	p.record(r, "apikeys.create", name, err)
	if err != nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, APIKeySecretResponse{Key: keyDTO(key), Secret: secret})
}

// handleRotate handles POST /api/v1/admin/api-keys/{id}/rotate.
//
//	@Summary		Rotate an API key
//	@Description	Issues a new secret for a key. Requires the apikeys.manage permission.
//	@Tags			apikeys
//	@ID				apikeys.rotate
//	@Produce		json
//	@Param			id	path	string	true	"key id"
//	@Success		200	{object}	APIKeySecretResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		404	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/api-keys/{id}/rotate [post]
func (p *Plugin) handleRotate(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	key, secret, err := p.store.Rotate(r.Context(), id)
	p.record(r, "apikeys.rotate", id.String(), err)
	if errors.Is(err, domain.ErrKeyNotFound) {
		httpapi.WriteError(w, r, errKeyNotFound)
		return
	}
	if err != nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, APIKeySecretResponse{Key: keyDTO(key), Secret: secret})
}

// handleRevoke handles DELETE /api/v1/admin/api-keys/{id}.
//
//	@Summary		Revoke an API key
//	@Description	Deletes an API key immediately. Requires the apikeys.manage permission.
//	@Tags			apikeys
//	@ID				apikeys.revoke
//	@Param			id	path	string	true	"key id"
//	@Success		204	"Revoked"
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		422	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/admin/api-keys/{id} [delete]
func (p *Plugin) handleRevoke(w http.ResponseWriter, r *http.Request) {
	if p.store == nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrUnprocessable)
		return
	}
	err = p.store.Revoke(r.Context(), id)
	p.record(r, "apikeys.revoke", id.String(), err)
	if err != nil {
		httpapi.WriteError(w, r, errKeysUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func keyDTO(key domain.APIKey) APIKey {
	dto := APIKey{
		ID:          key.ID.String(),
		Name:        key.Name,
		KeyPrefix:   key.KeyPrefix,
		Permissions: key.Permissions,
		CreatedAt:   key.CreatedAt.UTC().Format(time.RFC3339),
	}
	if key.LastUsedAt != nil {
		dto.LastUsedAt = key.LastUsedAt.UTC().Format(time.RFC3339)
	}
	return dto
}

func cleanPermissions(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func (p *Plugin) record(r *http.Request, action, target string, err error) {
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
		Permission:     string(PermissionManage),
		TargetType:     "api_key",
		TargetID:       target,
		Result:         result,
		RequestID:      httpapi.RequestIDFromContext(r.Context()),
	})
}
