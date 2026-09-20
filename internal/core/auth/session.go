package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrSessionNotFound is returned when a session token is unknown.
	ErrSessionNotFound = errors.New("auth: session not found")
	// ErrSessionExpired is returned when a session has passed its expiry.
	ErrSessionExpired = errors.New("auth: session expired")
)

// Session is a server-side session record.
//
// ID is an opaque, high-entropy random string. It is stored as the cookie
// value; PostgreSQL-backed stores are expected to persist only a hash of it
// once the identity plugin is implemented.
type Session struct {
	ID        string
	UserID    uuid.UUID
	DiscordID string
	Roles     []string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Store persists sessions server-side.
type Store interface {
	Create(ctx context.Context, session Session) error
	Get(ctx context.Context, id string) (Session, error)
	Delete(ctx context.Context, id string) error
}

// ManagerOptions configures a Manager.
type ManagerOptions struct {
	Store      Store
	TTL        time.Duration
	Secure     bool
	CookieName string
	SameSite   http.SameSite
	Now        func() time.Time
	// HashToken derives the value persisted server-side from the opaque cookie
	// token. It defaults to SHA-256 hex.
	HashToken func(string) string
	// RoleSource, when set, refreshes the principal's roles on every resolve so
	// role changes take effect without a re-login.
	RoleSource RoleSource
}

// Manager creates, resolves and revokes server-side sessions.
type Manager struct {
	store      Store
	ttl        time.Duration
	secure     bool
	cookieName string
	sameSite   http.SameSite
	now        func() time.Time
	hash       func(string) string
	roles      RoleSource
}

// NewManager creates a session manager.
func NewManager(opts ManagerOptions) *Manager {
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	cookieName := opts.CookieName
	if cookieName == "" {
		cookieName = "acgw_session"
	}
	sameSite := opts.SameSite
	if sameSite == http.SameSiteDefaultMode {
		sameSite = http.SameSiteLaxMode
	}
	hash := opts.HashToken
	if hash == nil {
		hash = sha256Hex
	}
	return &Manager{
		store:      opts.Store,
		ttl:        opts.TTL,
		secure:     opts.Secure,
		cookieName: cookieName,
		sameSite:   sameSite,
		now:        now,
		hash:       hash,
		roles:      opts.RoleSource,
	}
}

// ParseSameSite maps a configuration string to an http.SameSite value.
func ParseSameSite(value string) (http.SameSite, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "lax":
		return http.SameSiteLaxMode, nil
	case "strict":
		return http.SameSiteStrictMode, nil
	case "none":
		return http.SameSiteNoneMode, nil
	default:
		return http.SameSiteDefaultMode, fmt.Errorf("auth: invalid SameSite value %q", value)
	}
}

// TTL returns the configured session lifetime.
func (m *Manager) TTL() time.Duration {
	return m.ttl
}

// Create starts a new session for the principal and returns the opaque token.
func (m *Manager) Create(ctx context.Context, p Principal) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	now := m.now()
	session := Session{
		ID:        m.hash(token),
		UserID:    p.UserID,
		DiscordID: p.DiscordID,
		Roles:     append([]string(nil), p.Roles...),
		CreatedAt: now,
		ExpiresAt: now.Add(m.ttl),
	}
	if err := m.store.Create(ctx, session); err != nil {
		return "", fmt.Errorf("auth: create session: %w", err)
	}
	return token, nil
}

// Resolve validates a token and returns the associated principal.
func (m *Manager) Resolve(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrSessionNotFound
	}
	session, err := m.store.Get(ctx, m.hash(token))
	if err != nil {
		return Principal{}, err
	}
	if !m.now().Before(session.ExpiresAt) {
		if err := m.store.Delete(ctx, m.hash(token)); err != nil {
			return Principal{}, fmt.Errorf("%w: cleanup failed: %v", ErrSessionExpired, err)
		}
		return Principal{}, ErrSessionExpired
	}
	roles := session.Roles
	if m.roles != nil {
		if fresh, err := m.roles.RolesForUser(ctx, session.UserID); err == nil {
			roles = fresh
		}
	}
	return Principal{
		UserID:    session.UserID,
		DiscordID: session.DiscordID,
		Roles:     append([]string(nil), roles...),
	}, nil
}

// Revoke deletes a session. Unknown tokens are not an error.
func (m *Manager) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return m.store.Delete(ctx, m.hash(token))
}

// TokenFromRequest reads the session token from the request cookie.
func (m *Manager) TokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// SetCookie writes the session cookie. Max-Age is derived from the manager
// clock so injected clocks and clock skew behave consistently.
func (m *Manager) SetCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(expires.Sub(m.now()).Seconds()),
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: m.sameSite,
	})
}

// ClearCookie expires the session cookie.
func (m *Manager) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: m.sameSite,
	})
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: random token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
