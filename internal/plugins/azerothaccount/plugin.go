// Package azerothaccount owns account lifecycle capabilities (create, change
// password, set email), the account listing and the community user <-> account
// links, plus the AzerothCore CLI syntax required to implement them.
package azerothaccount

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/userdir"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/domain"
)

// Name is the stable plugin name.
const Name = "azeroth-account"

// Command names owned by this plugin.
const (
	CommandCreateAccount    = "account.create"
	CommandChangePassword   = "account.change-password"
	CommandSetEmail         = "account.set-email"
	CommandResolveAccount   = "account.resolve"
	ServiceAccountDirectory = "azeroth.account.directory"
)

// identityUserDirectory is the cross-plugin capability published by
// identity-discord.
const identityUserDirectory = "identity.user.directory"

// AccountDirectory is the synchronous capability published by this plugin.
type AccountDirectory interface {
	LinkedAccount(ctx context.Context, userID string) (username string, accountID *int64, err error)
}

// LinkStore persists community user <-> AzerothCore account links.
type LinkStore interface {
	Upsert(ctx context.Context, userID uuid.UUID, accountUsername string, accountID *int64) (domain.Link, error)
	Get(ctx context.Context, userID uuid.UUID) (domain.Link, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	List(ctx context.Context) ([]domain.Link, error)
}

// Config configures the azeroth-account plugin.
type Config struct {
	Executor azerothcore.CommandExecutor
	Accounts azerothdb.AccountReader
	Links    LinkStore
	Audit    audit.Recorder
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	executor azerothcore.CommandExecutor
	accounts azerothdb.AccountReader
	links    LinkStore
	audit    audit.Recorder
	users    userdir.Directory
}

// New creates the azeroth-account plugin. Accounts, Links and Audit may be nil;
// the corresponding features then report a specific unavailability error.
func New(cfg Config) *Plugin {
	recorder := cfg.Audit
	if recorder == nil {
		recorder = audit.NopRecorder{}
	}
	return &Plugin{
		executor: cfg.Executor,
		accounts: cfg.Accounts,
		links:    cfg.Links,
		audit:    recorder,
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

	directory, err := services.Consume[userdir.Directory](reg.Services, identityUserDirectory)
	if err != nil {
		return fmt.Errorf("azeroth-account: user directory unavailable: %w", err)
	}
	p.users = directory

	if err := commands.RegisterTyped(reg.Commands, CommandCreateAccount, p.createAccount); err != nil {
		return err
	}
	if err := commands.RegisterTyped(reg.Commands, CommandChangePassword, p.changePassword); err != nil {
		return err
	}
	if err := commands.RegisterTyped(reg.Commands, CommandSetEmail, p.setEmail); err != nil {
		return err
	}

	reg.Mux.Handle("POST /api/v1/azeroth/accounts",
		reg.RequirePermission(PermissionAccountManage, http.HandlerFunc(p.handleCreateAccount)))
	reg.Mux.Handle("POST /api/v1/azeroth/accounts/{username}/password",
		reg.RequirePermission(PermissionAccountManage, http.HandlerFunc(p.handleChangePassword)))
	reg.Mux.Handle("PUT /api/v1/azeroth/accounts/{username}/email",
		reg.RequirePermission(PermissionAccountManage, http.HandlerFunc(p.handleSetEmail)))
	reg.Mux.Handle("GET /api/v1/azeroth/accounts",
		reg.RequirePermission(PermissionAccountList, http.HandlerFunc(p.handleListAccounts)))

	reg.Mux.Handle("POST /api/v1/azeroth/account-links",
		reg.RequirePermission(PermissionAccountLink, http.HandlerFunc(p.handleCreateLink)))
	reg.Mux.Handle("GET /api/v1/azeroth/account-links",
		reg.RequirePermission(PermissionAccountRead, http.HandlerFunc(p.handleListLinks)))
	reg.Mux.Handle("GET /api/v1/azeroth/account-links/{user_id}",
		reg.RequirePermission(PermissionAccountRead, http.HandlerFunc(p.handleGetLink)))
	reg.Mux.Handle("DELETE /api/v1/azeroth/account-links/{user_id}",
		reg.RequirePermission(PermissionAccountLink, http.HandlerFunc(p.handleDeleteLink)))

	reg.Mux.Handle("GET /api/v1/azeroth/me/account",
		reg.RequirePermission(PermissionAccountSelf, http.HandlerFunc(p.handleMyAccount)))
	reg.Mux.Handle("POST /api/v1/azeroth/me/account",
		rateLimit(reg, reg.RequirePermission(PermissionAccountSelf,
			http.HandlerFunc(p.handleCreateMyAccount))))

	return services.Provide[AccountDirectory](reg.Services, ServiceAccountDirectory, p)
}

// CreateAccountRequest is the typed input of account.create.
type CreateAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
} // @name CreateAccountRequest

// ChangePasswordRequest is the typed input of account.change-password.
type ChangePasswordRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
} // @name ChangePasswordRequest

// SetEmailRequest is the typed input of account.set-email.
type SetEmailRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
} // @name SetEmailRequest

var (
	// ErrMissingUsername is returned when a request has no account username.
	ErrMissingUsername = errors.New("azeroth-account: username is required")
	// ErrMissingPassword is returned when a request has no password.
	ErrMissingPassword = errors.New("azeroth-account: password is required")
)

func (p *Plugin) createAccount(ctx context.Context, req CreateAccountRequest) (string, error) {
	username, err := azerothcore.SafeIdentifier(req.Username)
	if err != nil {
		return "", err
	}
	password, err := azerothcore.SafeAccountPassword(req.Password)
	if err != nil {
		return "", err
	}
	command := fmt.Sprintf(".account create %s %s", username, password)
	if req.Email != "" {
		email, err := azerothcore.SafeEmail(req.Email)
		if err != nil {
			return "", err
		}
		command = fmt.Sprintf(".account create %s %s %s", username, password, email)
	}
	return p.executor.Execute(ctx, command)
}

func (p *Plugin) changePassword(ctx context.Context, req ChangePasswordRequest) (string, error) {
	username, err := azerothcore.SafeIdentifier(req.Username)
	if err != nil {
		return "", err
	}
	password, err := azerothcore.SafeAccountPassword(req.Password)
	if err != nil {
		return "", err
	}
	command := fmt.Sprintf(".account set password %s %s %s", username, password, password)
	return p.executor.Execute(ctx, command)
}

func (p *Plugin) setEmail(ctx context.Context, req SetEmailRequest) (string, error) {
	username, err := azerothcore.SafeIdentifier(req.Username)
	if err != nil {
		return "", err
	}
	email, err := azerothcore.SafeEmail(req.Email)
	if err != nil {
		return "", err
	}
	command := fmt.Sprintf(".account set email %s %s", username, email)
	return p.executor.Execute(ctx, command)
}

// LinkedAccount implements AccountDirectory.
func (p *Plugin) LinkedAccount(ctx context.Context, userID string) (string, *int64, error) {
	if p.links == nil {
		return "", nil, errors.New("azeroth-account: link store unavailable")
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", nil, fmt.Errorf("azeroth-account: invalid user id: %w", err)
	}
	link, err := p.links.Get(ctx, id)
	if err != nil {
		return "", nil, err
	}
	return link.AccountUsername, link.AccountID, nil
}
