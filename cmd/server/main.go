// Command server starts the ac-community-gw HTTP gateway.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/mconcepcionb/ac-community-gw/internal/adapters/azerothmysql"
	"github.com/mconcepcionb/ac-community-gw/internal/adapters/azerothsoap"
	"github.com/mconcepcionb/ac-community-gw/internal/adapters/discord"
	"github.com/mconcepcionb/ac-community-gw/internal/adapters/postgres"
	"github.com/mconcepcionb/ac-community-gw/internal/core/audit"
	"github.com/mconcepcionb/ac-community-gw/internal/core/auth"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothcore"
	"github.com/mconcepcionb/ac-community-gw/internal/core/azerothdb"
	"github.com/mconcepcionb/ac-community-gw/internal/core/commands"
	"github.com/mconcepcionb/ac-community-gw/internal/core/config"
	"github.com/mconcepcionb/ac-community-gw/internal/core/events"
	"github.com/mconcepcionb/ac-community-gw/internal/core/httpapi"
	"github.com/mconcepcionb/ac-community-gw/internal/core/logging"
	"github.com/mconcepcionb/ac-community-gw/internal/core/metrics"
	"github.com/mconcepcionb/ac-community-gw/internal/core/permissions"
	"github.com/mconcepcionb/ac-community-gw/internal/core/persistence"
	"github.com/mconcepcionb/ac-community-gw/internal/core/plugins"
	"github.com/mconcepcionb/ac-community-gw/internal/core/services"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount"
	azerothaccountrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothaccount/repository"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothadmin"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothcharacter"
	azerothcharacterrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothcharacter/repository"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothinfo"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothitem"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothstore"
	azerothstorerepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/azerothstore/repository"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/identitydiscord"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/identitydiscord/repository"
	"github.com/mconcepcionb/ac-community-gw/internal/plugins/reports"
	reportsrepo "github.com/mconcepcionb/ac-community-gw/internal/plugins/reports/repository"
)

// @title			ac-community-gw API
// @version		1.0
// @description	REST gateway between community applications and AzerothCore.
// @description	It hides the AzerothCore SOAP/CLI internals behind a typed, authenticated and authorized HTTP API.
func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := healthcheck(); err != nil {
			fmt.Fprintln(os.Stderr, "healthcheck:", err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ac-community-gw:", err)
		os.Exit(1)
	}
}

// healthcheck probes the local /healthz endpoint. It is used by the container
// HEALTHCHECK because the distroless image has no shell or HTTP client.
func healthcheck() error {
	addr := os.Getenv("ACGW_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if host == "" || host == "::" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(host, port)+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel, cfg.LogFormat, os.Stdout)
	if cfg.Metrics.Enabled && cfg.Metrics.Token == "" {
		logger.Warn("metrics endpoint is enabled without ACGW_METRICS_TOKEN; it is reachable unauthenticated")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	readiness := persistence.NewRegistry()
	database, err := openDatabase(ctx, cfg, readiness, logger)
	if err != nil {
		return err
	}
	if database != nil {
		defer func() { _ = database.Close() }()
	}

	executor := azerothcore.GuardExecutor(newExecutor(cfg))

	sessionStore := auth.Store(auth.NewMemoryStore())
	var identityRepo *repository.Store
	var accountLinks azerothaccount.LinkStore
	var accountClaims azerothaccount.ClaimStore
	var storeRepo azerothstore.Store
	var characterVisibility azerothcharacter.VisibilityStore
	var reportsStore reports.Store
	if database != nil {
		identityRepo = repository.New(database.SQL())
		sessionStore = identityRepo
		accountRepo := azerothaccountrepo.New(database.SQL())
		accountLinks = accountRepo
		accountClaims = accountRepo
		storeRepo = azerothstorerepo.New(database.SQL())
		characterVisibility = azerothcharacterrepo.New(database.SQL())
		reportsStore = reportsrepo.New(database.SQL())
	}
	if storeRepo != nil {
		go runStoreReconciliation(ctx, storeRepo, logger)
	}

	sameSite, err := auth.ParseSameSite(cfg.Session.SameSite)
	if err != nil {
		return err
	}
	var roleSource auth.RoleSource
	if identityRepo != nil {
		roleSource = auth.NewCachingRoleSource(roleSourceFunc(identityRepo.UserRoles), 30*time.Second)
	}
	sessions := auth.NewManager(auth.ManagerOptions{
		Store:      sessionStore,
		TTL:        cfg.Session.TTL,
		Secure:     cfg.Session.Secure,
		CookieName: cfg.Session.CookieName,
		SameSite:   sameSite,
		RoleSource: roleSource,
	})

	provider, err := newDiscordProvider(cfg)
	if err != nil {
		return err
	}

	accountReader, closeAccountReader, err := newAccountReader(cfg, logger)
	if err != nil {
		return err
	}
	if closeAccountReader != nil {
		defer closeAccountReader()
	}

	characterReader, closeCharacters, err := newCharacterReader(cfg, logger)
	if err != nil {
		return err
	}
	if closeCharacters != nil {
		defer closeCharacters()
	}

	itemReader, closeItems, err := newItemReader(cfg, logger)
	if err != nil {
		return err
	}
	if closeItems != nil {
		defer closeItems()
	}

	metricRegistry := metrics.NewRegistry()
	trustedProxies, err := config.ParseTrustedProxies(cfg.Auth.TrustedProxies)
	if err != nil {
		return err
	}
	rateLimiter := httpapi.NewRateLimiterWithProxies(
		cfg.Auth.RateLimitPerMinute, cfg.Auth.RateLimitBurst, trustedProxies)

	commandRegistry := commands.NewRegistry()
	serviceRegistry := services.NewRegistry()
	eventBus := events.NewBus()
	permissionRegistry := permissions.NewRegistry()
	authorizer := permissions.NewAuthorizer()
	logRecorder := audit.NewLogRecorder(logger)
	var auditRecorder audit.Recorder = logRecorder
	var auditReader audit.Reader
	if database != nil {
		auditStore := postgres.NewAuditStore(database.SQL())
		auditRecorder = audit.NewMultiRecorder(logRecorder, auditStore)
		auditReader = auditStore
	}

	server := httpapi.New(httpapi.Dependencies{
		Config:      cfg,
		Logger:      logger,
		Sessions:    sessions,
		Permissions: permissionRegistry,
		Authorizer:  authorizer,
		Audit:       auditRecorder,
		Readiness:   readiness,
		Metrics:     metricRegistry,
	})

	registry := &plugins.Registry{
		Mux:               server.Mux(),
		Commands:          commandRegistry,
		Services:          serviceRegistry,
		Events:            eventBus,
		Permissions:       permissionRegistry,
		Audit:             auditRecorder,
		RequireAuth:       server.RequireAuth,
		RequirePermission: server.RequirePermission,
		RateLimit:         rateLimiter.Middleware,
	}

	manager := plugins.NewManager()
	manager.Add(identitydiscord.New(identitydiscord.Config{
		Sessions:          sessions,
		Provider:          provider,
		States:            newStateStore(identityRepo),
		Repository:        identityRepository(identityRepo),
		Authorizer:        authorizer,
		Events:            eventBus,
		Audit:             auditRecorder,
		Logger:            logger,
		Metrics:           metricRegistry,
		GuildID:           cfg.Discord.GuildID,
		RoleMappings:      cfg.Discord.RoleMappings,
		PostLoginRedirect: cfg.Discord.PostLoginRedirectURL,
	}))
	manager.Add(azerothaccount.New(azerothaccount.Config{
		Executor: executor,
		Accounts: accountReader,
		Links:    accountLinks,
		Claims:   accountClaims,
		Audit:    auditRecorder,
	}))
	manager.Add(azerothcharacter.New(azerothcharacter.Config{
		Executor:     executor,
		Characters:   characterReader,
		Accounts:     accountReader,
		Audit:        auditRecorder,
		Visibility:   characterVisibility,
		NoticeItemID: cfg.Notice.ItemID,
	}))
	manager.Add(azerothadmin.New(executor,
		azerothadmin.WithAudit(auditRecorder),
		azerothadmin.WithAuditReader(auditReader)))
	manager.Add(azerothinfo.New(executor))
	manager.Add(azerothitem.New(azerothitem.Config{Items: itemReader}))
	manager.Add(azerothstore.New(azerothstore.Config{
		Store: storeRepo,
		Audit: auditRecorder,
	}))
	manager.Add(reports.New(reports.Config{
		Store: reportsStore,
		Audit: auditRecorder,
	}))

	if err := manager.RegisterAll(ctx, registry); err != nil {
		return err
	}

	if identityRepo != nil {
		if err := identityRepo.SyncPermissions(ctx, permissionRegistry.Definitions()); err != nil {
			logger.Warn("permissions: sync to database failed", "error", err)
		}
	}

	if identityRepo != nil {
		roles, grants, err := loadAuthorizer(ctx, identityRepo, authorizer)
		if err != nil {
			return err
		}
		logger.Info("permissions: loaded role grants", "roles", roles, "grants", grants)
		go identitydiscord.RunCleanup(ctx, cfg.Session.CleanupInterval, logger, identityRepo)
		go runPermissionRefresh(ctx, identityRepo, authorizer, cfg.Permissions.RefreshInterval, logger)
	}

	logger.Info("ac-community-gw started",
		"env", cfg.Env,
		"plugins", manager.Names(),
		"commands", commandRegistry.Names(),
		"services", serviceRegistry.Names(),
		"permissions", len(permissionRegistry.Definitions()),
	)

	return server.Run(ctx)
}

func openDatabase(ctx context.Context, cfg *config.Config, readiness *persistence.Registry, logger *slog.Logger) (*postgres.DB, error) {
	if cfg.Database.URL == "" {
		logger.Warn("database URL is not configured; readiness will report unavailable")
		readiness.Add(postgres.Unavailable{Err: errors.New("database URL is not configured")})
		return nil, nil
	}

	database, err := postgres.Open(ctx, cfg.Database)
	if err != nil {
		logger.Warn("database connection failed; readiness will report unavailable", "error", err)
		readiness.Add(postgres.Unavailable{Err: err})
		return nil, nil
	}
	readiness.Add(database)

	if cfg.Database.AutoMigrate {
		migrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := postgres.Migrate(migrateCtx, database, "up"); err != nil {
			return nil, err
		}
	}
	return database, nil
}

func newExecutor(cfg *config.Config) azerothcore.CommandExecutor {
	if !cfg.AzerothSOAPConfigured() {
		return azerothcore.Unavailable{}
	}
	client, err := azerothsoap.New(azerothsoap.Config{
		URL:      cfg.Azeroth.SOAPURL,
		Username: cfg.Azeroth.SOAPUsername,
		Password: cfg.Azeroth.SOAPPassword,
		Timeout:  cfg.Azeroth.SOAPTimeout,
	})
	if err != nil {
		return azerothcore.Unavailable{}
	}
	return client
}

func newDiscordProvider(cfg *config.Config) (auth.DiscordProvider, error) {
	if !cfg.DiscordConfigured() {
		return nil, nil
	}
	return discord.New(discord.Config{
		ClientID:     cfg.Discord.ClientID,
		ClientSecret: cfg.Discord.ClientSecret,
		RedirectURL:  cfg.Discord.RedirectURL,
		AuthorizeURL: cfg.Discord.AuthorizeURL,
		TokenURL:     cfg.Discord.TokenURL,
		APIBaseURL:   cfg.Discord.APIBaseURL,
		Scopes:       cfg.Discord.Scopes,
		Timeout:      cfg.Discord.Timeout,
	})
}

func newAccountReader(cfg *config.Config, logger *slog.Logger) (azerothdb.AccountReader, func(), error) {
	if cfg.Azeroth.LoginDBDSN == "" {
		return nil, nil, nil
	}
	client, err := azerothmysql.New(azerothmysql.Config{DSN: cfg.Azeroth.LoginDBDSN})
	if err != nil {
		logger.Warn("azeroth login database unavailable; account listing will report an error", "error", err)
		return azerothdb.Unavailable{Err: err}, nil, nil
	}
	logger.Info("azeroth login database connected")
	return client, func() { _ = client.Close() }, nil
}

func newCharacterReader(cfg *config.Config, logger *slog.Logger) (azerothdb.CharacterReader, func(), error) {
	if cfg.Azeroth.CharacterDBDSN == "" {
		return nil, nil, nil
	}
	store, err := azerothmysql.NewCharacterStore(azerothmysql.Config{DSN: cfg.Azeroth.CharacterDBDSN})
	if err != nil {
		logger.Warn("azeroth character database unavailable; character reads will report an error", "error", err)
		return azerothdb.UnavailableCharacters{Err: err}, nil, nil
	}
	logger.Info("azeroth character database connected")
	return store, func() { _ = store.Close() }, nil
}

func newItemReader(cfg *config.Config, logger *slog.Logger) (azerothdb.ItemReader, func(), error) {
	if cfg.Azeroth.WorldDBDSN == "" {
		return nil, nil, nil
	}
	store, err := azerothmysql.NewItemStore(azerothmysql.Config{DSN: cfg.Azeroth.WorldDBDSN})
	if err != nil {
		logger.Warn("azeroth world database unavailable; item catalog will report an error", "error", err)
		return azerothdb.UnavailableItems{Err: err}, nil, nil
	}
	logger.Info("azeroth world database connected")
	return store, func() { _ = store.Close() }, nil
}

func newStateStore(store *repository.Store) identitydiscord.StateStore {
	if store == nil {
		return identitydiscord.NewMemoryStateStore()
	}
	return store
}

func identityRepository(store *repository.Store) identitydiscord.Repository {
	if store == nil {
		return nil
	}
	return store
}

// runStoreReconciliation periodically completes pending orders whose delivery
// output is already known.
func runStoreReconciliation(ctx context.Context, store azerothstore.Store, logger *slog.Logger) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			completed, err := store.ReconcilePendingOrders(ctx, 100)
			if err != nil {
				logger.Warn("store: reconciliation failed", "error", err)
				continue
			}
			if completed > 0 {
				logger.Info("store: reconciled pending orders", "completed", completed)
			}
		}
	}
}

// roleSourceFunc adapts a repository method to auth.RoleSource.
type roleSourceFunc func(context.Context, uuid.UUID) ([]string, error)

// RolesForUser implements auth.RoleSource.
func (f roleSourceFunc) RolesForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return f(ctx, userID)
}

func loadAuthorizer(ctx context.Context, repo *repository.Store, authorizer *permissions.Authorizer) (int, int, error) {
	grants, err := repo.ListRolePermissions(ctx)
	if err != nil {
		return 0, 0, err
	}
	byRole := make(map[permissions.Role][]permissions.Permission)
	for _, grant := range grants {
		role := permissions.Role(grant.Role)
		byRole[role] = append(byRole[role], permissions.Permission(grant.Permission))
	}
	authorizer.Replace(byRole)
	return len(byRole), len(grants), nil
}

// runPermissionRefresh reloads role -> permission grants from the database on an
// interval so grants changed in the database take effect without a restart.
func runPermissionRefresh(ctx context.Context, repo *repository.Store, authorizer *permissions.Authorizer, interval time.Duration, logger *slog.Logger) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			roles, grants, err := loadAuthorizer(ctx, repo, authorizer)
			if err != nil {
				logger.Warn("permissions: refresh failed", "error", err)
				continue
			}
			logger.Debug("permissions: refreshed role grants", "roles", roles, "grants", grants)
		}
	}
}
