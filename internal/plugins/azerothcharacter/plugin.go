// Package azerothcharacter owns AzerothCore character reads and the delivery of
// in-game mail, items and money.
package azerothcharacter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/delivery"
	"github.com/mconcepcionb/ac-community-gw/internal/core/notice"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/core/ttlcache"
)

// Name is the stable plugin name.
const Name = "azeroth-character"

// accountDirectory is the cross-plugin capability published by azeroth-account.
const accountDirectoryService = "azeroth.account.directory"

// itemCatalogService is the cross-plugin item lookup published by azeroth-item.
const itemCatalogService = "azeroth.item.catalog"

type accountDirectory interface {
	LinkedAccount(ctx context.Context, userID string) (username string, accountID *int64, err error)
}

// Config configures the azeroth-character plugin.
type Config struct {
	Executor     azerothcore.CommandExecutor
	Characters   azerothdb.CharacterReader
	Accounts     azerothdb.AccountReader
	Audit        audit.Recorder
	Visibility   VisibilityStore
	NoticeItemID int
}

// Plugin implements plugins.Plugin.
type Plugin struct {
	executor     azerothcore.CommandExecutor
	characters   azerothdb.CharacterReader
	accounts     azerothdb.AccountReader
	audit        audit.Recorder
	directory    accountDirectory
	visibility   VisibilityStore
	noticeItemID int
	boardCache   *ttlcache.Cache[LeaderboardResponse]
	registry     *services.Registry
}

// New creates the azeroth-character plugin.
func New(cfg Config) *Plugin {
	recorder := cfg.Audit
	if recorder == nil {
		recorder = audit.NopRecorder{}
	}
	return &Plugin{
		executor:     cfg.Executor,
		characters:   cfg.Characters,
		accounts:     cfg.Accounts,
		audit:        recorder,
		visibility:   cfg.Visibility,
		noticeItemID: cfg.NoticeItemID,
		boardCache:   ttlcache.New[LeaderboardResponse](publicBoardCacheTTL),
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

	directory, err := services.Consume[accountDirectory](reg.Services, accountDirectoryService)
	if err != nil {
		return fmt.Errorf("azeroth-character: account directory unavailable: %w", err)
	}
	p.directory = directory
	p.registry = reg.Services

	reg.Mux.Handle("GET /api/v1/azeroth/characters",
		reg.RequirePermission(PermissionCharacterList, http.HandlerFunc(p.handleListCharacters)))
	reg.Mux.Handle("GET /api/v1/azeroth/characters/{name}",
		reg.RequirePermission(PermissionCharacterList, http.HandlerFunc(p.handleGetCharacter)))
	reg.Mux.Handle("GET /api/v1/azeroth/users/{user_id}/characters",
		reg.RequirePermission(PermissionCharacterList, http.HandlerFunc(p.handleListUserCharacters)))
	reg.Mux.Handle("POST /api/v1/azeroth/mail",
		reg.RequirePermission(PermissionMailSend, http.HandlerFunc(p.handleSendMail)))
	reg.Mux.Handle("POST /api/v1/admin/characters/{name}/mail",
		reg.RequirePermission(PermissionAdminMailSend, http.HandlerFunc(p.handleAdminMail)))
	reg.Mux.Handle("GET /api/v1/azeroth/me/characters",
		reg.RequirePermission(PermissionCharacterSelf, http.HandlerFunc(p.handleMyCharacters)))
	reg.Mux.Handle("POST /api/v1/azeroth/me/mail",
		reg.RequirePermission(PermissionMailSelf, http.HandlerFunc(p.handleMyMail)))
	reg.Mux.Handle("GET /api/v1/azeroth/me/characters/visibility",
		reg.RequirePermission(PermissionCharacterSelf, http.HandlerFunc(p.handleListVisibility)))
	reg.Mux.Handle("PUT /api/v1/azeroth/me/characters/{name}/visibility",
		reg.RequirePermission(PermissionCharacterSelf, http.HandlerFunc(p.handleSetVisibility)))
	reg.Mux.Handle("GET /api/v1/azeroth/leaderboards/{board}",
		reg.RequirePermission(PermissionLeaderboardRead, http.HandlerFunc(p.handleLeaderboard)))
	reg.Mux.Handle("GET /api/v1/public/leaderboards/{board}",
		rateLimit(reg, http.HandlerFunc(p.handlePublicLeaderboard)))
	if err := services.Provide[Directory](reg.Services, DirectoryService, p); err != nil {
		return err
	}
	if err := services.Provide[notice.Service](reg.Services, NoticeService, p); err != nil {
		return err
	}
	return services.Provide[delivery.Service](reg.Services, DeliveryService, p)
}
