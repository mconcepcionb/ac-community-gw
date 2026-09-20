// Package azerothinfo exposes read-only AzerothCore information and publishes
// a synchronous capability other plugins may consume.
package azerothinfo

import (
	"context"
	"net/http"
	"time"

	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/ttlcache"
)

// publicCacheTTL bounds how stale a public status response may be.
const publicCacheTTL = 15 * time.Second

// Name is the stable plugin name.
const Name = "azeroth-info"

// ServiceServerInfo is the capability published by this plugin.
const ServiceServerInfo = "azeroth.info.server"

// ServerInfo is the result of a server status query.
type ServerInfo struct {
	Output            string
	Version           string
	ConnectedPlayers  int
	CharactersInWorld int
	ConnectionPeak    int
	Queue             int
	Uptime            string
}

// ServerInfoService is the synchronous capability contract. Consumers declare
// this interface in their own package and resolve it by name from the registry.
type ServerInfoService interface {
	Status(ctx context.Context) (ServerInfo, error)
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	executor    azerothcore.CommandExecutor
	statusCache *ttlcache.Cache[StatusResponse]
}

// New creates the azeroth-info plugin.
func New(executor azerothcore.CommandExecutor) *Plugin {
	return &Plugin{executor: executor, statusCache: ttlcache.New[StatusResponse](publicCacheTTL)}
}

// Name implements plugins.Plugin.
func (p *Plugin) Name() string { return Name }

// Register implements plugins.Plugin.
func (p *Plugin) Register(_ context.Context, reg *plugins.Registry) error {
	for _, def := range permissionDefs() {
		if err := reg.Permissions.Register(def); err != nil {
			return err
		}
	}
	if err := services.Provide[ServerInfoService](reg.Services, ServiceServerInfo, p); err != nil {
		return err
	}
	reg.Mux.Handle("GET /api/v1/azeroth/info/status",
		reg.RequirePermission(PermissionInfoPublicRead, http.HandlerFunc(p.handleStatus)))
	reg.Mux.Handle("GET /api/v1/public/status",
		rateLimit(reg, http.HandlerFunc(p.handlePublicStatus)))
	return nil
}

// handlePublicStatus handles GET /api/v1/public/status.
//
//	@Summary		Public server status
//	@Description	Reports connected players, peak, queue and uptime without authentication. Cached briefly and rate-limited per IP.
//	@Tags			azeroth-info
//	@ID				azeroth.public.status
//	@Produce		json
//	@Success		200	{object}	StatusResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/public/status [get]
func (p *Plugin) handlePublicStatus(w http.ResponseWriter, r *http.Request) {
	if cached, ok := p.statusCache.Get("status"); ok {
		httpapi.WriteJSON(w, http.StatusOK, cached)
		return
	}
	info, err := p.Status(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrBadGateway)
		return
	}
	response := StatusResponse{
		Output:            info.Output,
		Version:           info.Version,
		ConnectedPlayers:  info.ConnectedPlayers,
		CharactersInWorld: info.CharactersInWorld,
		ConnectionPeak:    info.ConnectionPeak,
		Queue:             info.Queue,
		Uptime:            info.Uptime,
	}
	p.statusCache.Set("status", response)
	httpapi.WriteJSON(w, http.StatusOK, response)
}

func rateLimit(reg *plugins.Registry, next http.Handler) http.Handler {
	if reg.RateLimit == nil {
		return next
	}
	return reg.RateLimit(next)
}

// Status implements ServerInfoService.
func (p *Plugin) Status(ctx context.Context) (ServerInfo, error) {
	output, err := p.executor.Execute(ctx, ".server info")
	if err != nil {
		return ServerInfo{}, err
	}
	return parseServerInfo(output), nil
}

// StatusResponse is the body returned by GET /api/v1/azeroth/info/status.
type StatusResponse struct {
	Output            string `json:"output"`
	Version           string `json:"version"`
	ConnectedPlayers  int    `json:"connected_players"`
	CharactersInWorld int    `json:"characters_in_world"`
	ConnectionPeak    int    `json:"connection_peak"`
	Queue             int    `json:"queue"`
	Uptime            string `json:"uptime"`
} // @name AzerothStatusResponse

// handleStatus reports the AzerothCore server status.
//
//	@Summary		AzerothCore server status
//	@Description	Reports connected players, peak, queue and uptime. Requires the azeroth.info.public.read permission.
//	@Tags			azeroth-info
//	@ID				azeroth.info.status
//	@Produce		json
//	@Success		200	{object}	StatusResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		403	{object}	httpapi.ErrorResponse
//	@Failure		502	{object}	httpapi.ErrorResponse
//	@Failure		503	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/azeroth/info/status [get]
func (p *Plugin) handleStatus(w http.ResponseWriter, r *http.Request) {
	info, err := p.Status(r.Context())
	if err != nil {
		httpapi.WriteError(w, r, httpapi.ErrBadGateway)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, StatusResponse{
		Output:            info.Output,
		Version:           info.Version,
		ConnectedPlayers:  info.ConnectedPlayers,
		CharactersInWorld: info.CharactersInWorld,
		ConnectionPeak:    info.ConnectionPeak,
		Queue:             info.Queue,
		Uptime:            info.Uptime,
	})
}
