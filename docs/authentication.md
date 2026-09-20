# Authentication

## Identity source

Discord is the source of truth for user authentication. The stable identity key
is `discord_user_id`; usernames are never used as identity.

Authentication (who the user is) and authorization (what the user may do) are
separate concerns: Discord authenticates, the gateway authorizes.

## Flow

OAuth2 Authorization Code flow via `identity-discord`:

```
GET  /api/v1/auth/discord/login      redirect to Discord
GET  /api/v1/auth/discord/callback   exchange code, provision user, create session
POST /api/v1/auth/logout             revoke session
GET  /api/v1/me                      current principal
```

The login endpoint mints a single-use `state` and a PKCE (`S256`) verifier and
challenge. The callback consumes the state, exchanges the code, provisions the
user, synchronises roles, creates a session and redirects to
`ACGW_DISCORD_POST_LOGIN_REDIRECT_URL` when configured. Discord access tokens
live only for the duration of the callback; they are never persisted or logged.

The Discord HTTP client lives in `internal/adapters/discord` and is injected
through the core `auth.DiscordProvider` interface, mirroring the SOAP adapter.

Degradation is explicit: when Discord credentials are missing the auth endpoints
return `503 discord_not_configured`; when the session/identity store is
unavailable they return `503 identity_storage_unavailable`. The gateway itself
keeps running.

See [runbooks/discord-oauth-setup.md](runbooks/discord-oauth-setup.md).

## Sessions

Server-side sessions are used. JWT is deliberately not introduced.

- random opaque session ID (32 bytes, hex)
- persisted server-side in PostgreSQL; only the SHA-256 hash of the token is
  stored (in-memory store in local development without a database)
- `HttpOnly` cookie
- `Secure` cookie (configurable; must be `true` in production)
- `SameSite=Lax`
- expiry (`ACGW_SESSION_TTL`); expired sessions are cleaned periodically
- a fresh token is minted on every login (fixation defence)
- roles are refreshed from the identity store on every resolve, behind a short
  cache (default 30s), so a role change is effective without a re-login

The cookie contains only the opaque token, never personal data.

## Provisioning

On first successful login the gateway upserts a `community_users` row and a
`discord_identities` row. The `identity-discord` plugin owns those tables. The
upsert is keyed on `discord_id`, so concurrent first logins create exactly one
user. Events `identity.user_provisioned` and `identity.user_authenticated` are
published after the transaction commits.

## Roles

Discord guild roles are mapped to internal roles either through the
`discord_role_mappings` table (owned by `identity-discord`) or the
`ACGW_DISCORD_ROLE_MAPPINGS` configuration; both are merged. Only mapped roles
reach the principal and the session; unmapped Discord roles grant nothing.
Role -> permission grants are loaded from `role_permissions` into the local
authorizer at startup. See [permissions.md](permissions.md).

## Hardening

- `/api/v1/auth/discord/login` and `/api/v1/auth/discord/callback` are
  rate-limited per client IP (`ACGW_AUTH_RATE_LIMIT_PER_MINUTE`,
  `ACGW_AUTH_RATE_LIMIT_BURST`); rejected requests return `429` with a
  `Retry-After` header.
- Cookie `SameSite` is configurable (`ACGW_SESSION_COOKIE_SAMESITE`, default
  `lax`); `none` requires `Secure`.
- Startup fails when `ACGW_ENV=production` and the session cookie is not
  `Secure` or the redirect URL is not HTTPS.
- Optional counters are exposed at `GET /metrics` when
  `ACGW_METRICS_ENABLED=true`, protected by `ACGW_METRICS_TOKEN` when set.

## See also

- [ADR 0004](ADR/0004-server-side-sessions.md)
- [security.md](security.md)
