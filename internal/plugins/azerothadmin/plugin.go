// Package azerothadmin owns administrative capabilities over AzerothCore
// accounts and the AzerothCore CLI syntax required to implement them.
package azerothadmin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/storeview"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
)

// Name is the stable plugin name.
const Name = "azeroth-admin"

// Command names owned by this plugin.
const (
	CommandBanAccount   = "account.ban"
	CommandUnbanAccount = "account.unban"
	CommandSetGMLevel   = "account.set-gmlevel"
)

// Capabilities consumed from other plugins.
const (
	accountDirectoryService   = "azeroth.account.directory"
	identityUserAdminService  = "identity.user.admin"
	characterDirectoryService = "azeroth.character.directory"
	storeAccountService       = "azeroth.store.account"
)

// accountDirectory is the capability published by azeroth-account.
type accountDirectory interface {
	LinkedAccount(ctx context.Context, userID string) (username string, accountID *int64, err error)
}

// userAdmin is the staff-facing capability published by identity-discord.
type userAdmin interface {
	UserByID(ctx context.Context, userID uuid.UUID) (userdir.User, error)
	Roles(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// characterDirectory is the capability published by azeroth-character.
type characterDirectory interface {
	CharactersByUser(ctx context.Context, userID string) ([]azerothdb.Character, error)
}

// storeAccount is the capability published by azeroth-store.
type storeAccount interface {
	Wallet(ctx context.Context, userID uuid.UUID) (int64, error)
	Orders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]storeview.Order, error)
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	executor    azerothcore.CommandExecutor
	audit       audit.Recorder
	auditReader audit.Reader
	// registry resolves cross-plugin capabilities on demand so the aggregate
	// does not depend on plugin registration order.
	registry *services.Registry
}

// Option customizes a Plugin at construction time.
type Option func(*Plugin)

// WithAudit injects the audit recorder.
func WithAudit(recorder audit.Recorder) Option {
	return func(p *Plugin) {
		if recorder != nil {
			p.audit = recorder
		}
	}
}

// WithAuditReader injects the audit reader used by the audit viewer.
func WithAuditReader(reader audit.Reader) Option {
	return func(p *Plugin) {
		if reader != nil {
			p.auditReader = reader
		}
	}
}

// New creates the azeroth-admin plugin.
func New(executor azerothcore.CommandExecutor, opts ...Option) *Plugin {
	p := &Plugin{executor: executor, audit: audit.NopRecorder{}}
	for _, opt := range opts {
		opt(p)
	}
	return p
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

	p.registry = reg.Services

	if err := commands.RegisterTyped(reg.Commands, CommandBanAccount, p.banAccount); err != nil {
		return err
	}
	if err := commands.RegisterTyped(reg.Commands, CommandUnbanAccount, p.unbanAccount); err != nil {
		return err
	}
	if err := commands.RegisterTyped(reg.Commands, CommandSetGMLevel, p.setGMLevel); err != nil {
		return err
	}

	reg.Mux.Handle("POST /api/v1/azeroth/accounts/{username}/ban",
		reg.RequirePermission(PermissionAdminAccountsBan, http.HandlerFunc(p.handleBanAccount)))
	reg.Mux.Handle("POST /api/v1/azeroth/accounts/{username}/unban",
		reg.RequirePermission(PermissionAdminAccountsBan, http.HandlerFunc(p.handleUnbanAccount)))
	reg.Mux.Handle("PUT /api/v1/azeroth/accounts/{username}/gmlevel",
		reg.RequirePermission(PermissionAdminAccountsGMLevel, http.HandlerFunc(p.handleSetGMLevel)))

	reg.Mux.Handle("GET /api/v1/azeroth/online",
		reg.RequirePermission(PermissionAdminPlayersRead, http.HandlerFunc(p.handleOnlineList)))
	reg.Mux.Handle("POST /api/v1/azeroth/players/{name}/kick",
		reg.RequirePermission(PermissionAdminPlayersKick, http.HandlerFunc(p.handleKick)))
	reg.Mux.Handle("POST /api/v1/azeroth/players/{name}/mute",
		reg.RequirePermission(PermissionAdminPlayersMute, http.HandlerFunc(p.handleMute)))
	reg.Mux.Handle("POST /api/v1/azeroth/players/{name}/unmute",
		reg.RequirePermission(PermissionAdminPlayersMute, http.HandlerFunc(p.handleUnmute)))
	reg.Mux.Handle("POST /api/v1/azeroth/characters/{name}/ban",
		reg.RequirePermission(PermissionAdminCharactersBan, http.HandlerFunc(p.handleBanCharacter)))
	reg.Mux.Handle("POST /api/v1/azeroth/characters/{name}/unban",
		reg.RequirePermission(PermissionAdminCharactersBan, http.HandlerFunc(p.handleUnbanCharacter)))
	reg.Mux.Handle("POST /api/v1/azeroth/announce",
		reg.RequirePermission(PermissionAdminAnnounce, http.HandlerFunc(p.handleAnnounce)))
	reg.Mux.Handle("GET /api/v1/admin/users/{id}",
		reg.RequirePermission(PermissionAdminUsersRead, http.HandlerFunc(p.handleUser360)))
	reg.Mux.Handle("GET /api/v1/admin/audit",
		reg.RequirePermission(PermissionAdminAuditRead, http.HandlerFunc(p.handleAuditLog)))
	return nil
}

// BanAccountRequest is the typed input of account.ban.
type BanAccountRequest struct {
	Username string `json:"username"`
	Duration string `json:"duration"`
	Reason   string `json:"reason"`
} // @name BanAccountRequest

// UnbanAccountRequest is the typed input of account.unban.
type UnbanAccountRequest struct {
	Username string `json:"username"`
} // @name UnbanAccountRequest

// SetGMLevelRequest is the typed input of account.set-gmlevel.
type SetGMLevelRequest struct {
	Username string `json:"username"`
	Level    int    `json:"level"`
	Realm    string `json:"realm"`
} // @name SetGMLevelRequest

// ErrMissingUsername is returned when a request has no account username.
var ErrMissingUsername = errors.New("azeroth-admin: username is required")

// Validation errors for administrative command inputs.
var (
	// ErrInvalidDuration is returned for a malformed ban duration.
	ErrInvalidDuration = errors.New("azeroth-admin: invalid duration")
	// ErrInvalidRealm is returned when a realm value is not an integer.
	ErrInvalidRealm = errors.New("azeroth-admin: invalid realm")
	// ErrInvalidGMLevel is returned when a GM level is out of range.
	ErrInvalidGMLevel = errors.New("azeroth-admin: invalid GM level")
)

const (
	minGMLevel = 0
	maxGMLevel = 4
)

func (p *Plugin) banAccount(ctx context.Context, req BanAccountRequest) (string, error) {
	username, err := azerothcore.SafeIdentifier(req.Username)
	if err != nil {
		return "", err
	}
	duration := req.Duration
	if duration == "" {
		duration = "1d"
	}
	duration, err = azerothcore.SafeBanDuration(duration)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidDuration, err)
	}
	reason := azerothcore.SafeQuotedText(req.Reason)
	if reason == "" {
		reason = "No reason"
	}
	command := fmt.Sprintf(".ban account %s %s %s", username, duration, reason)
	return p.executor.Execute(ctx, command)
}

func (p *Plugin) unbanAccount(ctx context.Context, req UnbanAccountRequest) (string, error) {
	username, err := azerothcore.SafeIdentifier(req.Username)
	if err != nil {
		return "", err
	}
	command := fmt.Sprintf(".unban account %s", username)
	return p.executor.Execute(ctx, command)
}

func (p *Plugin) setGMLevel(ctx context.Context, req SetGMLevelRequest) (string, error) {
	username, err := azerothcore.SafeIdentifier(req.Username)
	if err != nil {
		return "", err
	}
	if req.Level < minGMLevel || req.Level > maxGMLevel {
		return "", fmt.Errorf("%w: must be between %d and %d", ErrInvalidGMLevel, minGMLevel, maxGMLevel)
	}
	realm := req.Realm
	if realm == "" {
		realm = "-1"
	}
	if _, err := strconv.Atoi(realm); err != nil {
		return "", fmt.Errorf("%w: must be an integer", ErrInvalidRealm)
	}
	command := fmt.Sprintf(".account set gmlevel %s %d %s", username, req.Level, realm)
	return p.executor.Execute(ctx, command)
}
