# Discord authentication

## Goal

Implement Discord OAuth2 authentication, durable server-side session
persistence, community user provisioning and Discord-role synchronisation.

Discord is the source of authentication. The stable identity key is
`discord_user_id`; usernames are never used as identity.

## Status

| Area | State |
| --- | --- |
| `identity-discord` plugin, routes, permissions, events | implemented |
| `auth.Manager` + hashed server-side sessions | implemented |
| Discord HTTP client | implemented |
| OAuth2 login/callback handlers (state + PKCE) | implemented |
| PostgreSQL session store (hashed tokens) | implemented |
| User provisioning | implemented |
| Discord role -> internal role mapping | implemented |
| `GET /me`, `POST /logout` | implemented |
| Integration tests against PostgreSQL | implemented |
| Rate limiting / HTTP hardening (ticket 005) | implemented |

## Context

- [authentication.md](../../authentication.md) — identity model.
- [security.md](../../security.md) — trust boundaries and secrets handling.
- [ADR 0004](../../ADR/0004-server-side-sessions.md) — server-side sessions.
- [ADR 0008](../../ADR/0008-no-cross-plugin-imports.md) — import boundaries.

## Scope

- OAuth2 Authorization Code flow (confidential client, `state` + PKCE).
- Discord user + guild-member-role fetch.
- PostgreSQL session store, token hashing, expiry and periodic cleanup.
- Idempotent provisioning of `community_users` / `discord_identities`.
- DB-driven mapping of Discord role id -> internal role, applied on login.
- Load role -> permission grants from PostgreSQL into the authorizer.
- Events and audit for login, provisioning and role changes.
- Local development through a `cloudflared` tunnel container.

## Out of scope

- Discord bot, slash commands, guild management.
- Granting permissions directly from unmapped Discord roles.
- Refresh-token storage or offline access.
- Front-end application (the gateway redirects to it after login).
- Multi-tenant / multiple guilds (one configured guild for now).

## Design

```
Browser            Gateway (identity-discord)              Discord
   |  GET /auth/discord/login                                  |
   |---------------------------------------------------------->|
   |  302 + Location: authorize?...state=..&code_challenge=..  |
   |  <--------------------------------------------------------|
   |  user consents                                            |
   |  GET /auth/discord/callback?code=..&state=..              |
   |---------------------------------------------------------->|
   |            lookup state, exchange code (+verifier)        |
   |            GET /users/@me                                 |
   |            GET /users/@me/guilds/{guild}/member           |
   |            upsert community_users + discord_identities     |
   |            map Discord roles -> internal roles             |
   |            create session (hashed token in PostgreSQL)     |
   |  <-- 302 to post-login URL, Set-Cookie: acgw_session       |
   |  GET /api/v1/me (cookie)                                   |
   |---------------------------------------------------------->|
   |            resolve session -> principal                    |
   |  <-- 200 {user_id, discord_id, roles}                      |
```

### Architecture constraint

The architecture test (`internal/architecture`) forbids a plugin from importing
`internal/adapters` and core from importing adapters. The Discord HTTP client
therefore follows the same pattern as the SOAP adapter:

- core defines the interface `auth.DiscordProvider` and value types;
- `internal/adapters/discord` implements it with `net/http`;
- `cmd/server` constructs the adapter and injects it into the plugin.

`identity-discord` only ever sees the core interface.

### Component map

| Component | Path |
| --- | --- |
| Provider contract + value types | `internal/core/auth/discord.go` |
| Discord HTTP client | `internal/adapters/discord/client.go` (+ `client_test.go`) |
| Plugin wiring | `internal/plugins/identitydiscord/plugin.go` |
| Login/callback/logout/me handlers | `internal/plugins/identitydiscord/oauth.go` |
| Provisioning | `internal/plugins/identitydiscord/provision.go` |
| Role mapping | `internal/plugins/identitydiscord/roles.go` |
| In-memory OAuth state store | `internal/plugins/identitydiscord/memory.go` |
| Expiry janitor | `internal/plugins/identitydiscord/janitor.go` |
| PostgreSQL stores | `internal/plugins/identitydiscord/repository/store.go` |
| SQL queries | `internal/plugins/identitydiscord/repository/queries.sql` |
| Schema | `migrations/00004_discord_oauth.sql` |
| Config | `internal/core/config/config.go` |
| Development test frontend | `web/index.html` (`ACGW_WEB_DIR`) |
| Wiring | `cmd/server/main.go` |
| Compose + tunnel | `compose.yaml` |

### Configuration

New variables (defaults shown):

```
ACGW_DISCORD_CLIENT_ID=
ACGW_DISCORD_CLIENT_SECRET=
ACGW_DISCORD_REDIRECT_URL=
ACGW_DISCORD_AUTHORIZE_URL=https://discord.com/oauth2/authorize
ACGW_DISCORD_TOKEN_URL=https://discord.com/api/oauth2/token
ACGW_DISCORD_API_BASE_URL=https://discord.com/api/v10
ACGW_DISCORD_SCOPES=identify guilds.members.read
ACGW_DISCORD_GUILD_ID=
ACGW_DISCORD_POST_LOGIN_REDIRECT_URL=
ACGW_DISCORD_TIMEOUT=10s
```

`DiscordConfigured()` (already present) stays the switch that enables the
Discord client. When Discord or the database is not configured the plugin
degrades gracefully: `/login` and `/callback` return the shared JSON error
envelope instead of panicking.

### Schema changes

One new migration, `migrations/00004_discord_oauth.sql`:

```sql
-- +goose Up

ALTER TABLE sessions
    ADD COLUMN roles text[] NOT NULL DEFAULT '{}';

ALTER TABLE community_users
    ADD COLUMN roles text[] NOT NULL DEFAULT '{}',
    ADD COLUMN roles_synced_at timestamptz;

CREATE TABLE discord_role_mappings (
    discord_role_id text PRIMARY KEY,
    role            text NOT NULL REFERENCES roles (name) ON DELETE CASCADE,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE oauth_states (
    state         text PRIMARY KEY,
    code_verifier text NOT NULL DEFAULT '',
    return_to     text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    expires_at    timestamptz NOT NULL
);

CREATE INDEX oauth_states_expires_at_idx ON oauth_states (expires_at);

-- +goose Down

DROP TABLE oauth_states;
DROP TABLE discord_role_mappings;
ALTER TABLE community_users DROP COLUMN roles_synced_at, DROP COLUMN roles;
ALTER TABLE sessions DROP COLUMN roles;
```

`discord_id` is not duplicated on `sessions`; the session store joins
`community_users` on `user_id`.

## Implementation phases

1. **Ticket 001 — OAuth flow.** Core provider contract, `internal/adapters/discord`
   client, `oauth_states` table + store, login/callback handlers, config.
2. **Ticket 002 — Session persistence.** `auth.Manager` token hashing, PostgreSQL
   session store, janitor, wiring in `cmd/server`.
3. **Ticket 003 — Provisioning.** Idempotent `community_users` upsert, identity
   upsert, events, audit.
4. **Ticket 004 — Role sync.** Mapping table + queries, role resolution on login,
   `DiscordRolesChanged` event, authorizer loading at startup.
5. **Ticket 005 — Hardening.** Rate limiting on the auth endpoints, configurable
   `SameSite`, production safety checks and optional `/metrics` counters.

All tickets are implemented. Integration tests live behind the `integration`
build tag and require `ACGW_DATABASE_URL`.

## Security

- `state` is single-use and short-lived (10 minutes); it is deleted before the
  code exchange to defeat replay.
- PKCE (`S256`) binds the authorization code to this client.
- Access/refresh tokens are held in memory for the duration of the callback only;
  they are never logged, persisted, put in events, the audit log or cookies.
- Session tokens are 32 random bytes, hex-encoded; only their SHA-256 hash is
  stored. A database dump never yields a usable session token.
- A new session token is minted on every login (fixation defence).
- `HttpOnly`, `SameSite=Lax`, `Secure` (configurable, required in production).
- Only mapped Discord roles influence authorisation; unmapped roles grant
  nothing and no `IsAdmin` shortcut exists.
- All Discord failures map to `502 bad_gateway`; no upstream body is echoed.

## Testing

Unit (no network, no PostgreSQL):

- adapter against `httptest` fake Discord token/user/member endpoints;
- handler tests with a fake `auth.DiscordProvider` and in-memory stores:
  state validation success/failure, upstream failure -> 502, session cookie set,
  `/me` returns the principal;
- provisioning idempotency with a fake repository;
- role mapping: mapped role grants, unmapped grants nothing, change event.

Integration (`//go:build integration`, disposable PostgreSQL):

- sessions survive restart, expiry rejects + deletes, logout revokes;
- concurrent first logins create exactly one `community_users` row;
- role change is persisted and reflected in the next session.

Architecture:

- `go test ./internal/architecture/...` must keep passing (no plugin -> adapter).

## Operations

- [Runbook: Discord OAuth setup](../../runbooks/discord-oauth-setup.md) —
  app creation, cloudflared tunnel, env, migration, end-to-end test.
- [Runbook: rotate Discord secret](../../runbooks/rotate-discord-secret.md).
- [Runbook: migrations](../../runbooks/migrations.md).

## Risks

| Risk | Mitigation |
| --- | --- |
| Token leakage via logs | never log token/code; redact in adapter errors |
| CSRF | single-use `state`, SameSite=Lax cookie |
| Session fixation | mint a fresh token on login |
| Multiple instances racing on first login | `ON CONFLICT (discord_id)` upsert |
| Unmapped Discord roles granting access | explicit mapping table only |
| Quick-tunnel URL churn breaking redirect URI | stable named tunnel for shared testing |
| Stale sessions/states growing unbounded | periodic janitor + indexes |

## Tickets

1. [001-discord-oauth-flow.md](ticket/001-discord-oauth-flow.md)
2. [002-session-persistence.md](ticket/002-session-persistence.md)
3. [003-user-provisioning.md](ticket/003-user-provisioning.md)
4. [004-role-synchronization.md](ticket/004-role-synchronization.md)
5. [005-hardening.md](ticket/005-hardening.md)
