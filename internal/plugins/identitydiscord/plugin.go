// Package identitydiscord owns Discord-based authentication and the community
// user/session persistence.
package identitydiscord

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/events"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/metrics"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	identitydiscordrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/identitydiscord/repository/generated"
)

// Name is the stable plugin name.
const Name = "identity-discord"

// ServiceUserDirectory is the capability published to resolve community users.
const ServiceUserDirectory = "identity.user.directory"

// ServiceUserAdmin is the staff-facing capability published to read community
// users and their roles.
const ServiceUserAdmin = "identity.user.admin"

// UserAdmin is the cross-plugin capability published by this plugin for staff
// reads.
type UserAdmin interface {
	// UserByID returns the community user's Discord profile.
	UserByID(ctx context.Context, userID uuid.UUID) (userdir.User, error)
	// Roles returns the internal roles assigned to the user.
	Roles(ctx context.Context, userID uuid.UUID) ([]string, error)
}

var (
	errDiscordNotConfigured = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"discord_not_configured", "Discord authentication is not configured")
	errIdentityStorageUnavailable = httpapi.NewAPIError(http.StatusServiceUnavailable,
		"identity_storage_unavailable", "identity storage is unavailable")
)

// unavailableError explains why the auth flow cannot run.
func (p *Plugin) unavailableError() error {
	if p.provider == nil {
		return errDiscordNotConfigured
	}
	return errIdentityStorageUnavailable
}

// StateStore persists short-lived OAuth states.
type StateStore interface {
	CreateOAuthState(ctx context.Context, state, codeVerifier, returnTo string, expiresAt time.Time) error
	ConsumeOAuthState(ctx context.Context, state string) (codeVerifier, returnTo string, err error)
}

// Repository persists community users and role mappings.
type Repository interface {
	Provision(ctx context.Context, user auth.DiscordUser) (userID uuid.UUID, created bool, err error)
	MappedRoles(ctx context.Context, discordRoleIDs []string) ([]string, error)
	UserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	UpdateUserRoles(ctx context.Context, userID uuid.UUID, roles []string) error
	UserIDByDiscordID(ctx context.Context, discordID string) (userID uuid.UUID, found bool, err error)
	ListUsers(ctx context.Context, filter string, limit, offset int) ([]userdir.User, error)
	ResolveUserByName(ctx context.Context, name string) ([]userdir.User, error)
	UserProfile(ctx context.Context, userID uuid.UUID) (userdir.User, error)
	// Role and permission administration.
	SyncPermissions(ctx context.Context, defs []permissions.Definition) error
	ListRolePermissions(ctx context.Context) ([]identitydiscordrepo.RolePermission, error)
	ListRoles(ctx context.Context) ([]string, error)
	UpsertRole(ctx context.Context, role string) error
	GrantRolePermission(ctx context.Context, role, permission string) error
	RevokeRolePermission(ctx context.Context, role, permission string) error
	ReplaceRolePermissions(ctx context.Context, role string, perms []string) error
	ListDiscordRoleMappings(ctx context.Context) ([]identitydiscordrepo.DiscordRoleMapping, error)
	UpsertDiscordRoleMapping(ctx context.Context, discordRoleID, role string) error
	DeleteDiscordRoleMapping(ctx context.Context, discordRoleID string) error
}

// Cleaner removes expired sessions and OAuth states.
type Cleaner interface {
	DeleteExpiredSessions(ctx context.Context) error
	DeleteExpiredOAuthStates(ctx context.Context) error
}

// PermissionLister resolves the permissions granted to a set of roles.
type PermissionLister interface {
	Permissions(roles []permissions.Role) []permissions.Permission
}

// Config configures the identity-discord plugin.
type Config struct {
	Sessions          *auth.Manager
	Provider          auth.DiscordProvider
	States            StateStore
	Repository        Repository
	Authorizer        PermissionLister
	Events            *events.Bus
	Audit             audit.Recorder
	Logger            *slog.Logger
	Metrics           *metrics.Registry
	GuildID           string
	RoleMappings      map[string]string
	PostLoginRedirect string
	StateTTL          time.Duration
	Now               func() time.Time
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	sessions          *auth.Manager
	provider          auth.DiscordProvider
	states            StateStore
	repo              Repository
	authorizer        PermissionLister
	events            *events.Bus
	audit             audit.Recorder
	logger            *slog.Logger
	metrics           *metrics.Registry
	guildID           string
	roleMappings      map[string]string
	postLoginRedirect string
	stateTTL          time.Duration
	now               func() time.Time
	permissions       *permissions.Registry
}

// New creates the identity-discord plugin.
func New(cfg Config) *Plugin {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	stateTTL := cfg.StateTTL
	if stateTTL <= 0 {
		stateTTL = 10 * time.Minute
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	auditRecorder := cfg.Audit
	if auditRecorder == nil {
		auditRecorder = audit.NopRecorder{}
	}
	return &Plugin{
		sessions:          cfg.Sessions,
		provider:          cfg.Provider,
		states:            cfg.States,
		repo:              cfg.Repository,
		authorizer:        cfg.Authorizer,
		events:            cfg.Events,
		audit:             auditRecorder,
		logger:            logger,
		metrics:           cfg.Metrics,
		guildID:           cfg.GuildID,
		roleMappings:      cfg.RoleMappings,
		postLoginRedirect: cfg.PostLoginRedirect,
		stateTTL:          stateTTL,
		now:               now,
	}
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

	p.permissions = reg.Permissions
	if err := services.Provide[userdir.Directory](reg.Services, ServiceUserDirectory, p); err != nil {
		return err
	}
	if err := services.Provide[UserAdmin](reg.Services, ServiceUserAdmin, p); err != nil {
		return err
	}

	reg.Mux.Handle("GET /api/v1/auth/discord/login", rateLimit(reg, http.HandlerFunc(p.handleLogin)))
	reg.Mux.Handle("GET /api/v1/auth/discord/callback", rateLimit(reg, http.HandlerFunc(p.handleCallback)))
	reg.Mux.HandleFunc("POST /api/v1/auth/logout", p.handleLogout)
	reg.Mux.Handle("GET /api/v1/me", reg.RequireAuth(http.HandlerFunc(p.handleMe)))
	reg.Mux.Handle("GET /api/v1/identity/users",
		reg.RequirePermission(PermissionUserRead, http.HandlerFunc(p.handleListUsers)))
	reg.Mux.Handle("GET /api/v1/admin/roles",
		reg.RequirePermission(PermissionRolesManage, http.HandlerFunc(p.handleListRoles)))
	reg.Mux.Handle("PUT /api/v1/admin/roles/{role}/permissions",
		reg.RequirePermission(PermissionRolesManage, http.HandlerFunc(p.handleReplaceRolePermissions)))
	reg.Mux.Handle("PUT /api/v1/admin/discord-role-mappings/{discord_role_id}",
		reg.RequirePermission(PermissionRolesManage, http.HandlerFunc(p.handleUpsertDiscordMapping)))
	reg.Mux.Handle("DELETE /api/v1/admin/discord-role-mappings/{discord_role_id}",
		reg.RequirePermission(PermissionRolesManage, http.HandlerFunc(p.handleDeleteDiscordMapping)))
	return nil
}

// ResolveUser implements UserDirectory.
func (p *Plugin) ResolveUser(ctx context.Context, discordID string) (uuid.UUID, bool, error) {
	if p.repo == nil {
		return uuid.Nil, false, errIdentityStorageUnavailable
	}
	return p.repo.UserIDByDiscordID(ctx, discordID)
}

// ResolveUserByName implements userdir.Directory. It returns exact
// (case-insensitive) matches on Discord username, global name or display name.
func (p *Plugin) ResolveUserByName(ctx context.Context, name string) ([]userdir.User, error) {
	if p.repo == nil {
		return nil, errIdentityStorageUnavailable
	}
	return p.repo.ResolveUserByName(ctx, name)
}

// ListUsers implements userdir.Directory.
func (p *Plugin) ListUsers(ctx context.Context, filter string, limit, offset int) ([]userdir.User, error) {
	if p.repo == nil {
		return nil, errIdentityStorageUnavailable
	}
	return p.repo.ListUsers(ctx, filter, limit, offset)
}

// UserByID implements UserAdmin.
func (p *Plugin) UserByID(ctx context.Context, userID uuid.UUID) (userdir.User, error) {
	if p.repo == nil {
		return userdir.User{}, errIdentityStorageUnavailable
	}
	return p.repo.UserProfile(ctx, userID)
}

// Roles implements UserAdmin.
func (p *Plugin) Roles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	if p.repo == nil {
		return nil, errIdentityStorageUnavailable
	}
	return p.repo.UserRoles(ctx, userID)
}

// rateLimit wraps a handler with the core rate limiter when one is provided.
func rateLimit(reg *plugins.Registry, next http.Handler) http.Handler {
	if reg.RateLimit == nil {
		return next
	}
	return reg.RateLimit(next)
}

// handleLogout revokes the current session and clears the cookie.
//
//	@Summary		Log out
//	@Description	Revokes the current server-side session and clears the session cookie.
//	@Tags			auth
//	@ID				auth.logout
//	@Success		204	"Session revoked"
//	@Failure		500	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/auth/logout [post]
func (p *Plugin) handleLogout(w http.ResponseWriter, r *http.Request) {
	if token := p.sessions.TokenFromRequest(r); token != "" {
		_ = p.sessions.Revoke(r.Context(), token)
	}
	p.sessions.ClearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// handleMe returns the authenticated principal.
//
//	@Summary		Current principal
//	@Description	Returns the authenticated community user with their Discord profile, internal roles and effective permissions.
//	@Tags			auth
//	@ID				auth.me
//	@Produce		json
//	@Success		200	{object}	MeResponse
//	@Failure		401	{object}	httpapi.ErrorResponse
//	@Failure		500	{object}	httpapi.ErrorResponse
//	@Router			/api/v1/me [get]
func (p *Plugin) handleMe(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, r, httpapi.ErrUnauthorized)
		return
	}
	response := MeResponse{
		UserID:      principal.UserID.String(),
		DiscordID:   principal.DiscordID,
		Roles:       rolesOrEmpty(principal.Roles),
		Permissions: p.effectivePermissions(principal.Roles),
	}
	if p.repo != nil {
		profile, err := p.repo.UserProfile(r.Context(), principal.UserID)
		if err != nil {
			p.logger.Warn("identity-discord: user profile unavailable", "error", err)
		} else {
			response.Username = profile.Username
			response.GlobalName = profile.GlobalName
			response.DisplayName = profile.DisplayName
			response.Avatar = profile.Avatar
			if !profile.CreatedAt.IsZero() {
				response.CreatedAt = profile.CreatedAt.UTC().Format(time.RFC3339)
			}
		}
	}
	httpapi.WriteJSON(w, http.StatusOK, response)
}
