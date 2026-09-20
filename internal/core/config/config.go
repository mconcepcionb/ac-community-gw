// Package config loads the gateway configuration from environment variables.
//
// Configuration is deliberately flat and explicit. Unknown variables are
// ignored; invalid values produce a startup error so misconfiguration is
// visible immediately.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the fully resolved runtime configuration.
type Config struct {
	Env             string
	HTTPAddr        string
	LogLevel        string
	LogFormat       string
	ShutdownTimeout time.Duration

	Database Database
	Discord  Discord
	Azeroth  Azeroth
	Session  Session
	Auth     Auth
	Metrics  Metrics
	// Permissions configures the role -> permission reload.
	Permissions Permissions
	// Notice configures in-game text notices.
	Notice Notice
}

// Database configures the gateway-owned PostgreSQL database.
type Database struct {
	URL             string
	AutoMigrate     bool
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Discord holds the OAuth2 credentials and endpoints for the Discord identity
// provider.
type Discord struct {
	ClientID             string
	ClientSecret         string
	RedirectURL          string
	AuthorizeURL         string
	TokenURL             string
	APIBaseURL           string
	Scopes               []string
	GuildID              string
	RoleMappings         map[string]string
	PostLoginRedirectURL string
	Timeout              time.Duration
}

// Azeroth holds the connection details for the AzerothCore SOAP endpoint and
// its login database.
type Azeroth struct {
	SOAPURL        string
	SOAPUsername   string
	SOAPPassword   string
	SOAPTimeout    time.Duration
	LoginDBDSN     string
	CharacterDBDSN string
	WorldDBDSN     string
}

// Session configures server-side session cookies.
type Session struct {
	CookieName      string
	TTL             time.Duration
	Secure          bool
	SameSite        string
	CleanupInterval time.Duration
}

// Auth configures authentication endpoint protection.
type Auth struct {
	RateLimitPerMinute int
	RateLimitBurst     int
	// TrustedProxies is a list of CIDRs (or bare IPs) whose X-Forwarded-For
	// headers are trusted. Empty means no proxy is trusted.
	TrustedProxies []string
}

// Metrics configures the optional counters endpoint.
type Metrics struct {
	Enabled bool
	Token   string
}

// Permissions configures how often role -> permission grants are reloaded from
// the database so changes take effect without a restart. A non-positive value
// disables the reload.
type Permissions struct {
	RefreshInterval time.Duration
}

// Notice configures in-game text notices. AzerothCore mail requires an
// enclosure, so a notice attaches Notice.ItemID; zero disables notices.
type Notice struct {
	ItemID int
}

// Load reads configuration from the process environment, applying defaults
// and validating the result.
func Load() (*Config, error) {
	roleMappings, err := envRoleMappings("ACGW_DISCORD_ROLE_MAPPINGS")
	if err != nil {
		return nil, err
	}
	p := &envParser{}
	cfg := &Config{
		Env:             env("ACGW_ENV", "development"),
		HTTPAddr:        env("ACGW_HTTP_ADDR", ":8080"),
		LogLevel:        env("ACGW_LOG_LEVEL", "info"),
		LogFormat:       env("ACGW_LOG_FORMAT", "json"),
		ShutdownTimeout: p.duration("ACGW_HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		Database: Database{
			URL:             env("ACGW_DATABASE_URL", ""),
			AutoMigrate:     p.boolean("ACGW_DB_AUTO_MIGRATE", false),
			MaxOpenConns:    p.integer("ACGW_DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    p.integer("ACGW_DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: p.duration("ACGW_DB_CONN_MAX_LIFETIME", time.Hour),
		},
		Discord: Discord{
			ClientID:             env("ACGW_DISCORD_CLIENT_ID", ""),
			ClientSecret:         env("ACGW_DISCORD_CLIENT_SECRET", ""),
			RedirectURL:          env("ACGW_DISCORD_REDIRECT_URL", ""),
			AuthorizeURL:         env("ACGW_DISCORD_AUTHORIZE_URL", "https://discord.com/oauth2/authorize"),
			TokenURL:             env("ACGW_DISCORD_TOKEN_URL", "https://discord.com/api/oauth2/token"),
			APIBaseURL:           env("ACGW_DISCORD_API_BASE_URL", "https://discord.com/api/v10"),
			Scopes:               envList("ACGW_DISCORD_SCOPES", "identify guilds.members.read"),
			GuildID:              env("ACGW_DISCORD_GUILD_ID", ""),
			RoleMappings:         roleMappings,
			PostLoginRedirectURL: env("ACGW_DISCORD_POST_LOGIN_REDIRECT_URL", ""),
			Timeout:              p.duration("ACGW_DISCORD_TIMEOUT", 10*time.Second),
		},
		Azeroth: Azeroth{
			SOAPURL:        env("ACGW_AZEROTH_SOAP_URL", ""),
			SOAPUsername:   env("ACGW_AZEROTH_SOAP_USERNAME", ""),
			SOAPPassword:   env("ACGW_AZEROTH_SOAP_PASSWORD", ""),
			SOAPTimeout:    p.duration("ACGW_AZEROTH_SOAP_TIMEOUT", 10*time.Second),
			LoginDBDSN:     env("ACGW_AZEROTH_LOGIN_DB_DSN", ""),
			CharacterDBDSN: env("ACGW_AZEROTH_CHARACTER_DB_DSN", ""),
			WorldDBDSN:     env("ACGW_AZEROTH_WORLD_DB_DSN", ""),
		},
		Session: Session{
			CookieName:      env("ACGW_SESSION_COOKIE_NAME", "acgw_session"),
			TTL:             p.duration("ACGW_SESSION_TTL", 24*time.Hour),
			Secure:          p.boolean("ACGW_SESSION_COOKIE_SECURE", true),
			SameSite:        env("ACGW_SESSION_COOKIE_SAMESITE", "lax"),
			CleanupInterval: p.duration("ACGW_SESSION_CLEANUP_INTERVAL", 15*time.Minute),
		},
		Auth: Auth{
			RateLimitPerMinute: p.integer("ACGW_AUTH_RATE_LIMIT_PER_MINUTE", 30),
			RateLimitBurst:     p.integer("ACGW_AUTH_RATE_LIMIT_BURST", 10),
			TrustedProxies:     envList("ACGW_TRUSTED_PROXIES", ""),
		},
		Metrics: Metrics{
			Enabled: p.boolean("ACGW_METRICS_ENABLED", false),
			Token:   env("ACGW_METRICS_TOKEN", ""),
		},
		Permissions: Permissions{
			RefreshInterval: p.duration("ACGW_PERMISSIONS_REFRESH_INTERVAL", 30*time.Second),
		},
		Notice: Notice{
			ItemID: p.integer("ACGW_NOTICE_ITEM_ID", 0),
		},
	}

	if err := p.err(); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// envParser accumulates errors from typed environment variables so Load can
// report every invalid value at once instead of silently falling back.
type envParser struct {
	errs []error
}

func (p *envParser) err() error {
	if len(p.errs) == 0 {
		return nil
	}
	return errors.Join(p.errs...)
}

func (p *envParser) boolean(key string, fallback bool) bool {
	value, err := envBool(key, fallback)
	if err != nil {
		p.errs = append(p.errs, err)
	}
	return value
}

func (p *envParser) integer(key string, fallback int) int {
	value, err := envInt(key, fallback)
	if err != nil {
		p.errs = append(p.errs, err)
	}
	return value
}

func (p *envParser) duration(key string, fallback time.Duration) time.Duration {
	value, err := envDuration(key, fallback)
	if err != nil {
		p.errs = append(p.errs, err)
	}
	return value
}

func (c *Config) validate() error {
	switch c.LogFormat {
	case "json", "text":
	default:
		return fmt.Errorf("config: ACGW_LOG_FORMAT must be json or text, got %q", c.LogFormat)
	}
	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "warning", "error":
	default:
		return fmt.Errorf("config: invalid ACGW_LOG_LEVEL %q", c.LogLevel)
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("config: ACGW_HTTP_ADDR must not be empty")
	}
	if c.Discord.RedirectURL != "" {
		if _, err := url.ParseRequestURI(c.Discord.RedirectURL); err != nil {
			return fmt.Errorf("config: invalid ACGW_DISCORD_REDIRECT_URL: %w", err)
		}
	}
	if c.Discord.PostLoginRedirectURL != "" {
		if _, err := url.ParseRequestURI(c.Discord.PostLoginRedirectURL); err != nil {
			return fmt.Errorf("config: invalid ACGW_DISCORD_POST_LOGIN_REDIRECT_URL: %w", err)
		}
	}
	if c.Discord.Timeout <= 0 {
		return fmt.Errorf("config: ACGW_DISCORD_TIMEOUT must be positive")
	}
	if c.Azeroth.SOAPURL != "" {
		if _, err := url.ParseRequestURI(c.Azeroth.SOAPURL); err != nil {
			return fmt.Errorf("config: invalid ACGW_AZEROTH_SOAP_URL: %w", err)
		}
	}
	if c.Session.TTL <= 0 {
		return fmt.Errorf("config: ACGW_SESSION_TTL must be positive")
	}
	if c.Session.CleanupInterval <= 0 {
		return fmt.Errorf("config: ACGW_SESSION_CLEANUP_INTERVAL must be positive")
	}
	switch strings.ToLower(c.Session.SameSite) {
	case "lax", "strict", "none":
	default:
		return fmt.Errorf("config: ACGW_SESSION_COOKIE_SAMESITE must be lax, strict or none, got %q", c.Session.SameSite)
	}
	if strings.EqualFold(c.Session.SameSite, "none") && !c.Session.Secure {
		return fmt.Errorf("config: ACGW_SESSION_COOKIE_SAMESITE=none requires ACGW_SESSION_COOKIE_SECURE=true")
	}
	if c.Auth.RateLimitPerMinute <= 0 {
		return fmt.Errorf("config: ACGW_AUTH_RATE_LIMIT_PER_MINUTE must be positive")
	}
	if c.Auth.RateLimitBurst <= 0 {
		return fmt.Errorf("config: ACGW_AUTH_RATE_LIMIT_BURST must be positive")
	}
	if _, err := ParseTrustedProxies(c.Auth.TrustedProxies); err != nil {
		return err
	}
	if c.Permissions.RefreshInterval < 0 {
		return fmt.Errorf("config: ACGW_PERMISSIONS_REFRESH_INTERVAL must not be negative")
	}
	if c.Notice.ItemID < 0 {
		return fmt.Errorf("config: ACGW_NOTICE_ITEM_ID must not be negative")
	}
	if strings.EqualFold(c.Env, "production") {
		if !c.Session.Secure {
			return fmt.Errorf("config: ACGW_SESSION_COOKIE_SECURE must be true when ACGW_ENV=production")
		}
		if c.Discord.RedirectURL != "" && !strings.HasPrefix(strings.ToLower(c.Discord.RedirectURL), "https://") {
			return fmt.Errorf("config: ACGW_DISCORD_REDIRECT_URL must use https when ACGW_ENV=production")
		}
		if c.Metrics.Enabled && c.Metrics.Token == "" {
			return fmt.Errorf("config: ACGW_METRICS_TOKEN is required when ACGW_METRICS_ENABLED=true and ACGW_ENV=production")
		}
		for name, value := range map[string]string{
			"ACGW_DISCORD_AUTHORIZE_URL": c.Discord.AuthorizeURL,
			"ACGW_DISCORD_TOKEN_URL":     c.Discord.TokenURL,
			"ACGW_DISCORD_API_BASE_URL":  c.Discord.APIBaseURL,
			"ACGW_AZEROTH_SOAP_URL":      c.Azeroth.SOAPURL,
		} {
			if !httpsOrLoopback(value) {
				return fmt.Errorf("config: %s must use https when ACGW_ENV=production", name)
			}
		}
	}
	return nil
}

// httpsOrLoopback reports whether a URL is HTTPS or a loopback HTTP endpoint.
// An empty value is allowed (the feature is simply disabled).
func httpsOrLoopback(value string) bool {
	if strings.TrimSpace(value) == "" {
		return true
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	if parsed.Scheme != "http" {
		return false
	}
	host := parsed.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// AzerothSOAPConfigured reports whether a SOAP endpoint is available.
func (c *Config) AzerothSOAPConfigured() bool {
	return c.Azeroth.SOAPURL != ""
}

// DiscordConfigured reports whether Discord OAuth2 credentials are present.
func (c *Config) DiscordConfigured() bool {
	return c.Discord.ClientID != "" && c.Discord.ClientSecret != "" && c.Discord.RedirectURL != ""
}

// ParseTrustedProxies parses a list of CIDRs or bare IP addresses into
// networks. Bare IPs become single-host networks.
func ParseTrustedProxies(entries []string) ([]*net.IPNet, error) {
	var networks []*net.IPNet
	for _, entry := range entries {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		if _, network, err := net.ParseCIDR(trimmed); err == nil {
			networks = append(networks, network)
			continue
		}
		ip := net.ParseIP(trimmed)
		if ip == nil {
			return nil, fmt.Errorf("config: invalid ACGW_TRUSTED_PROXIES entry %q", entry)
		}
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		networks = append(networks, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
	}
	return networks, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// envList reads a whitespace- or comma-separated list, dropping empty items.
func envList(key, fallback string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		raw = fallback
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// envRoleMappings reads "discord_role_id=internal_role" pairs separated by
// commas or whitespace, e.g. "1550848038216409149=ac-core.admin".
func envRoleMappings(key string) (map[string]string, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	mappings := make(map[string]string)
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	for _, field := range fields {
		discordRoleID, role, found := strings.Cut(field, "=")
		discordRoleID = strings.TrimSpace(discordRoleID)
		role = strings.TrimSpace(role)
		if !found || discordRoleID == "" || role == "" {
			return nil, fmt.Errorf("config: invalid %s entry %q, expected <discord_role_id>=<internal_role>", key, field)
		}
		mappings[discordRoleID] = role
	}
	return mappings, nil
}

func envBool(key string, fallback bool) (bool, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback, fmt.Errorf("config: invalid %s value %q", key, v)
	}
	return parsed, nil
}

func envInt(key string, fallback int) (int, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback, fmt.Errorf("config: invalid %s value %q", key, v)
	}
	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return fallback, fmt.Errorf("config: invalid %s value %q", key, v)
	}
	return parsed, nil
}
