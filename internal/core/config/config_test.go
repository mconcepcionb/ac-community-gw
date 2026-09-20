package config

import (
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Session.SameSite != "lax" {
		t.Fatalf("SameSite = %q", cfg.Session.SameSite)
	}
	if cfg.Session.CleanupInterval <= 0 {
		t.Fatalf("CleanupInterval = %v", cfg.Session.CleanupInterval)
	}
	if cfg.Auth.RateLimitPerMinute != 30 || cfg.Auth.RateLimitBurst != 10 {
		t.Fatalf("auth = %+v", cfg.Auth)
	}
	if cfg.Metrics.Enabled {
		t.Fatal("metrics should be disabled by default")
	}
}

func TestLoadProductionRequiresSecureCookie(t *testing.T) {
	t.Setenv("ACGW_ENV", "production")
	t.Setenv("ACGW_SESSION_COOKIE_SECURE", "false")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for insecure cookie in production")
	}
}

func TestLoadProductionRequiresHTTPSRedirect(t *testing.T) {
	t.Setenv("ACGW_ENV", "production")
	t.Setenv("ACGW_SESSION_COOKIE_SECURE", "true")
	t.Setenv("ACGW_DISCORD_REDIRECT_URL", "http://gateway.test/callback")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for http redirect in production")
	}
}

func TestLoadSameSiteNoneRequiresSecure(t *testing.T) {
	t.Setenv("ACGW_SESSION_COOKIE_SAMESITE", "none")
	t.Setenv("ACGW_SESSION_COOKIE_SECURE", "false")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for SameSite=none without Secure")
	}
}

func TestLoadRejectsInvalidSameSite(t *testing.T) {
	t.Setenv("ACGW_SESSION_COOKIE_SAMESITE", "bogus")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid SameSite")
	}
}

func TestLoadRejectsInvalidRateLimit(t *testing.T) {
	t.Setenv("ACGW_AUTH_RATE_LIMIT_PER_MINUTE", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-positive rate limit")
	}
}

func TestLoadRoleMappings(t *testing.T) {
	t.Setenv("ACGW_DISCORD_ROLE_MAPPINGS", "1550848038216409149=ac-core.admin, 42=member")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Discord.RoleMappings["1550848038216409149"] != "ac-core.admin" {
		t.Fatalf("mappings = %v", cfg.Discord.RoleMappings)
	}
	if cfg.Discord.RoleMappings["42"] != "member" {
		t.Fatalf("mappings = %v", cfg.Discord.RoleMappings)
	}
}

func TestLoadRejectsInvalidRoleMapping(t *testing.T) {
	t.Setenv("ACGW_DISCORD_ROLE_MAPPINGS", "not-a-pair")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid role mapping")
	}
}

func TestLoadTrustedProxies(t *testing.T) {
	t.Setenv("ACGW_TRUSTED_PROXIES", "10.0.0.0/8, 192.168.1.1")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	networks, err := ParseTrustedProxies(cfg.Auth.TrustedProxies)
	if err != nil {
		t.Fatalf("ParseTrustedProxies: %v", err)
	}
	if len(networks) != 2 {
		t.Fatalf("networks = %d, want 2", len(networks))
	}
}

func TestLoadRejectsInvalidTrustedProxy(t *testing.T) {
	t.Setenv("ACGW_TRUSTED_PROXIES", "not-a-cidr")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid trusted proxy")
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	t.Setenv("ACGW_SESSION_TTL", "25hours")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestLoadRejectsInvalidBool(t *testing.T) {
	t.Setenv("ACGW_METRICS_ENABLED", "treu")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid bool")
	}
}

func TestLoadReportsAllInvalidValues(t *testing.T) {
	t.Setenv("ACGW_SESSION_TTL", "nope")
	t.Setenv("ACGW_METRICS_ENABLED", "treu")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
	message := err.Error()
	if !strings.Contains(message, "ACGW_SESSION_TTL") || !strings.Contains(message, "ACGW_METRICS_ENABLED") {
		t.Fatalf("error did not name every variable: %v", err)
	}
}

func TestLoadAcceptsWarningLogLevel(t *testing.T) {
	t.Setenv("ACGW_LOG_LEVEL", "warning")
	if _, err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestLoadProductionMetricsRequiresToken(t *testing.T) {
	t.Setenv("ACGW_ENV", "production")
	t.Setenv("ACGW_SESSION_COOKIE_SECURE", "true")
	t.Setenv("ACGW_METRICS_ENABLED", "true")
	t.Setenv("ACGW_METRICS_TOKEN", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for metrics without token in production")
	}
}

func TestLoadDevelopmentMetricsWithoutTokenAllowed(t *testing.T) {
	t.Setenv("ACGW_ENV", "development")
	t.Setenv("ACGW_METRICS_ENABLED", "true")
	t.Setenv("ACGW_METRICS_TOKEN", "")
	if _, err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestLoadProductionRejectsCleartextUpstream(t *testing.T) {
	t.Setenv("ACGW_ENV", "production")
	t.Setenv("ACGW_SESSION_COOKIE_SECURE", "true")
	t.Setenv("ACGW_AZEROTH_SOAP_URL", "http://world.test/soap")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for cleartext SOAP url in production")
	}
}

func TestLoadProductionAllowsLoopbackHTTP(t *testing.T) {
	t.Setenv("ACGW_ENV", "production")
	t.Setenv("ACGW_SESSION_COOKIE_SECURE", "true")
	t.Setenv("ACGW_AZEROTH_SOAP_URL", "http://127.0.0.1:7878/")
	if _, err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}
