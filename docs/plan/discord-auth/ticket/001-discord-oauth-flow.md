# 001 — Discord OAuth2 flow

## Goal

Implement the Discord OAuth2 Authorization Code flow with `state` + PKCE, a
transport-only Discord client, and a short-lived server-side OAuth state store.

## Context

The `identity-discord` plugin currently stubs `/login` and `/callback`
(`plugin.go:48`). `config.Discord` already carries `ClientID`,
`ClientSecret` and `RedirectURL`. The architecture test forbids a plugin
importing `internal/adapters`, so the client is injected through a core
interface.

## Deliverables

### 1. Core provider contract — `internal/core/auth/discord.go`

```go
package auth

import (
	"context"
	"time"
)

// DiscordUser is the subset of the Discord user object the gateway needs.
type DiscordUser struct {
	ID         string
	Username   string
	GlobalName string
	Avatar     string
}

// DiscordToken is an OAuth2 token response.
//
// It is secret: never log, persist, publish or audit it.
type DiscordToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

// DiscordProvider is the transport contract implemented by
// internal/adapters/discord and consumed by the identity-discord plugin.
type DiscordProvider interface {
	AuthorizeURL(state, codeChallenge string) string
	Exchange(ctx context.Context, code, codeVerifier string) (DiscordToken, error)
	CurrentUser(ctx context.Context, accessToken string) (DiscordUser, error)
	GuildMemberRoleIDs(ctx context.Context, accessToken, guildID string) ([]string, error)
}
```

### 2. Adapter — `internal/adapters/discord/client.go`

`net/http` only (ADR 0002). `Config` carries `ClientID`, `ClientSecret`,
`RedirectURL`, `AuthorizeURL`, `TokenURL`, `APIBaseURL`, `Scopes`,
`GuildID`, `Timeout`, `HTTPClient` (injectable for tests).

- `AuthorizeURL(state, codeChallenge)` builds the query with `client_id`,
  `redirect_uri`, `response_type=code`, `scope`, `state`, `code_challenge`,
  `code_challenge_method=S256`, `prompt=consent`.
- `Exchange` POSTs `application/x-www-form-urlencoded` to the token endpoint
  with `client_id`, `client_secret`, `grant_type=authorization_code`, `code`,
  `redirect_uri`, `code_verifier`.
- `CurrentUser` GETs `/users/@me` with `Authorization: Bearer <token>`.
- `GuildMemberRoleIDs` GETs `/users/@me/guilds/{guildID}/member` and returns
  `roles`.
- Response bodies are `io.LimitReader`-bounded. Non-2xx and decode failures
  wrap `ErrUpstream`; the upstream body is never included in the error text.

Register `var _ auth.DiscordProvider = (*Client)(nil)`.

### 3. OAuth state store — `internal/plugins/identitydiscord/oauth_state.go`

Interface plus a PostgreSQL implementation backed by `oauth_states`
(migration in ticket 002 covers the table; add the queries here):

```sql
-- name: CreateOAuthState :exec
INSERT INTO oauth_states (state, code_verifier, return_to, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ConsumeOAuthState :one
DELETE FROM oauth_states
WHERE state = $1
RETURNING state, code_verifier, return_to, created_at, expires_at;

-- name: DeleteExpiredOAuthStates :exec
DELETE FROM oauth_states WHERE expires_at < now();
```

`Consume` deletes and returns in one statement, making `state` single-use even
under concurrency.

### 4. Handlers — `internal/plugins/identitydiscord/oauth.go`

- `handleLogin`: mint 32-byte `state`, mint a PKCE verifier (43–128 chars),
  derive `code_challenge = base64url(sha256(verifier))` without padding,
  persist `{state, verifier, return_to=redirect query param}`, then
  `302` to `provider.AuthorizeURL(state, challenge)`.
- `handleCallback`:
  1. reject a missing `code`/`state` with `ErrBadRequest`;
  2. `Consume` the state; reject unknown/expired with `ErrBadRequest`;
  3. call `Exchange(code, verifier)`; map failures to `ErrBadGateway`;
  4. call `CurrentUser`; map failures to `ErrBadGateway`;
  5. hand off to provisioning (ticket 003) and role sync (ticket 004);
  6. create the session, set the cookie, rotate (see below);
  7. `302` to `ACGW_DISCORD_POST_LOGIN_REDIRECT_URL` when set, otherwise
     `200 {user_id, discord_id, roles}`.
- `handleLogout` already revokes and clears the cookie; keep it.
- `handleMe` already returns the principal; keep it.

Rotation/fixation: before creating the new session, if the request carried a
valid session cookie, revoke that old session.

### 5. Config — `internal/core/config/config.go`

Extend `Discord` with `AuthorizeURL`, `TokenURL`, `APIBaseURL`, `Scopes`,
`GuildID`, `PostLoginRedirectURL`, `Timeout`, read from the env vars listed in
the plan README. Parse `Scopes` from a space-separated string. Validate that
`RedirectURL` and `PostLoginRedirectURL` parse when non-empty.

### 6. Wiring — `cmd/server/main.go`

```go
var provider auth.DiscordProvider
if cfg.DiscordConfigured() {
	provider, _ = discord.New(discord.Config{ /* from cfg.Discord */ })
}
manager.Add(identitydiscord.New(identitydiscord.Config{Provider: provider, ...}))
```

## Acceptance criteria

- `/login` redirects to Discord with `response_type=code`, the configured
  scopes, an unguessable `state` and an S256 `code_challenge`.
- `/callback` rejects a missing/unknown/expired `state` with `400 bad_request`.
- A failed token exchange or user fetch returns `502 bad_gateway`.
- A successful callback sets an `HttpOnly` session cookie and redirects to the
  post-login URL (or returns the principal).
- No token, code or secret is logged, stored in a cookie or put in an event.
- `go test ./...` and `go test ./internal/architecture/...` pass.

## Tests

- `internal/adapters/discord/client_test.go`: `httptest` fake token, user and
  member endpoints; assert form fields, bearer header, timeout and upstream
  error mapping.
- `internal/plugins/identitydiscord/oauth_test.go`: fake `DiscordProvider`
  (records `state`, verifier, challenge); success, invalid state, expired state,
  upstream failure; assert cookie attributes.
- State store test: `Consume` twice fails the second time.
