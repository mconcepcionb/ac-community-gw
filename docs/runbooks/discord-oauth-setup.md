# Runbook: set up Discord OAuth

## Purpose

Stand up Discord OAuth2 authentication for `ac-community-gw` locally (or in a
shared dev environment) end to end: create the Discord application, expose the
gateway over HTTPS with a `cloudflared` tunnel, configure the environment, apply
migrations, then verify login, sessions and role mapping.

This runbook assumes the code in `docs/plan/discord-auth/` is implemented. Until
then `/api/v1/auth/discord/login` and `/callback` return `501 not_implemented`.

## Preconditions

- Access to <https://discord.com/developers/applications>.
- A Discord server (guild) you can administer for testing roles.
- Docker + Docker Compose, or a host running `cloudflared`.
- Go 1.27+, Task, sqlc, Goose (for local, non-Docker runs).
- `.env` created from `.env.example` (never commit it).

## Part 1 — Create the Discord application

1. Open <https://discord.com/developers/applications> and choose **New
   Application**. Name it (e.g. `ac-community-gw-dev`) and accept the terms.
2. On **General Information**, copy the **Application ID**. This is
   `ACGW_DISCORD_CLIENT_ID`.
3. Open **OAuth2 -> General**:
   - **Client Secret** -> **Reset Secret**, copy it. This is
     `ACGW_DISCORD_CLIENT_SECRET`. Treat it as a password; it is shown once.
   - Under **Redirects**, add the callback URI(s). Discord requires HTTPS for
     non-localhost. Add both while developing:
     - `http://localhost:8080/api/v1/auth/discord/callback` (local, no tunnel)
     - `https://<your-tunnel-host>/api/v1/auth/discord/callback` (tunnel)
   - Save changes.
4. On **OAuth2 -> URL Generator** confirm the scopes the gateway will request:
   - `identify` — read the user's id, username, global name, avatar.
   - `guilds.members.read` — read the user's own member record in the guild,
     which is where the role ids come from.
   You do not need the `bot` scope. The gateway builds the authorize URL itself,
   so the generated URL is only a sanity check.
5. Get the **guild id**: in the Discord client enable **User Settings ->
   Advanced -> Developer Mode**, then right-click the test server ->
   **Copy Server ID**. This is `ACGW_DISCORD_GUILD_ID`.
6. Get the **role id(s)**: right-click a role (or a member's role) ->
   **Copy Role ID**. The `@everyone` role id equals the guild id; it grants
   nothing unless explicitly mapped.
7. Make sure the account you will log in with **is a member of that guild**;
   `guilds.members.read` only works for guilds the user belongs to.

## Part 2 — Expose the gateway with cloudflared

Discord rejects `http` redirect URIs (other than localhost), so a tunnel gives a
stable HTTPS origin. Two options.

### Option A — Quick tunnel (fastest, URL changes on restart)

Add this service to `compose.yaml`:

```yaml
  cloudflared:
    image: cloudflare/cloudflared:latest
    command: tunnel --no-autoupdate --url http://app:8080
    restart: unless-stopped
    depends_on:
      - app
```

Start it and read the assigned URL from its logs:

```bash
docker compose up -d
docker compose logs -f cloudflared
# look for: https://<random>.trycloudflare.com
```

Copy that origin and register
`https://<random>.trycloudflare.com/api/v1/auth/discord/callback` in the Discord
**Redirects** list, and set `ACGW_DISCORD_REDIRECT_URL` to the same value. Every
container restart changes the URL, so repeat this step.

### Option B — Named tunnel (stable host, recommended for shared testing)

Requires a Cloudflare account and a domain. Create a tunnel in the Cloudflare
dashboard (Zero Trust -> Networks -> Tunnels), point a hostname such as
`acgw-dev.example.com` at `http://app:8080`, and copy the tunnel token:

```yaml
  cloudflared:
    image: cloudflare/cloudflared:latest
    command: tunnel --no-autoupdate run --token ${CLOUDFLARE_TUNNEL_TOKEN}
    restart: unless-stopped
    depends_on:
      - app
```

Then `ACGW_DISCORD_REDIRECT_URL=https://acgw-dev.example.com/api/v1/auth/discord/callback`.
The URL is stable, so the Discord redirect entry does not change.

> The tunnel publishes only the gateway. Never expose the AzerothCore SOAP
> endpoint through it (see `docs/security.md`).

## Part 3 — Configure the environment

Append to `.env` (and keep `.env.example` in sync with empty values):

```dotenv
# Discord OAuth2
ACGW_DISCORD_CLIENT_ID=<application id>
ACGW_DISCORD_CLIENT_SECRET=<client secret>
ACGW_DISCORD_REDIRECT_URL=https://<tunnel-host>/api/v1/auth/discord/callback
ACGW_DISCORD_GUILD_ID=<guild id>
ACGW_DISCORD_POST_LOGIN_REDIRECT_URL=http://localhost:8080/

# The SPA is served by Vite (`task dev`, on :5173) or by Caddy (`task docker:up`,
# on :8080), never by the gateway itself.

# Optional overrides (defaults shown)
# ACGW_DISCORD_AUTHORIZE_URL=https://discord.com/oauth2/authorize
# ACGW_DISCORD_TOKEN_URL=https://discord.com/api/oauth2/token
# ACGW_DISCORD_API_BASE_URL=https://discord.com/api/v10
# ACGW_DISCORD_SCOPES=identify guilds.members.read
# ACGW_DISCORD_TIMEOUT=10s

# Sessions (Secure must be true when served over HTTPS via the tunnel)
ACGW_SESSION_COOKIE_SECURE=true
# ACGW_SESSION_COOKIE_SAMESITE=lax
# ACGW_SESSION_CLEANUP_INTERVAL=15m

# Optional: auth rate limiting and counters
# ACGW_AUTH_RATE_LIMIT_PER_MINUTE=30
# ACGW_AUTH_RATE_LIMIT_BURST=10
# ACGW_METRICS_ENABLED=true
# ACGW_METRICS_TOKEN=<secret>
```

Verify the gateway reports Discord as configured in the startup log. When
`ACGW_DISCORD_CLIENT_ID`, `ACGW_DISCORD_CLIENT_SECRET` or
`ACGW_DISCORD_REDIRECT_URL` is empty, the Discord client is disabled and the
auth endpoints return an error envelope instead of redirecting.

## Part 4 — Migrate and run

```bash
task db:up
task db:migrate          # applies 00004_discord_oauth.sql
task db:status
task run                 # or: task docker:up --build
```

Expected startup log: plugins include `identity-discord`; the session store is
PostgreSQL when `ACGW_DATABASE_URL` is set. `GET /readyz` must return `200`.

## Part 5 — Map Discord roles (optional but needed for authorization)

Run against the gateway database. Replace the ids with yours:

```sql
INSERT INTO roles (name, description)
VALUES ('member', 'Community member'),
       ('moderator', 'Community moderator')
ON CONFLICT (name) DO NOTHING;

INSERT INTO discord_role_mappings (discord_role_id, role)
VALUES ('<discord role id>', 'member'),
       ('<discord role id>', 'moderator')
ON CONFLICT (discord_role_id) DO UPDATE SET role = EXCLUDED.role;

-- GET /api/v1/me only requires an authenticated session, so member/moderator
-- need no permission for the profile page. Grant gw.* or azeroth.* permissions
-- here as your community requires, for example:
--   INSERT INTO role_permissions (role, permission)
--   VALUES ('member', 'gw.store.catalog.read')
--   ON CONFLICT DO NOTHING;
```

Permission names must exist in the registry, otherwise the authorization check
fails closed. Restart the gateway so `role_permissions` load into the
authorizer. An unmapped Discord role contributes nothing.

### Demo admin role (for the test frontend)

To run the AzerothCore commands from the test frontend, grant a role the
`azeroth.*` and `gw.*` permissions and map it to a Discord role. The helper script inserts
the permissions, creates an internal role with all of them and maps your Discord
role id:

```bash
# host psql
psql "$ACGW_DATABASE_URL" \
  -v discord_role_id=<role id> -v internal_role=ac-core.admin \
  -f scripts/seed_demo_admin.sql

# or via the compose postgres
Get-Content scripts/seed_demo_admin.sql -Raw |
  docker compose exec -T postgres psql -U acgw -d acgw \
    -v discord_role_id=<role id> -v internal_role=ac-core.admin
```

Then log out and back in so the new roles are part of the session, and restart
the gateway so it loads the grants.

### Static role mappings (alternative to the table)

Mappings can also be provided through configuration and are merged with the
`discord_role_mappings` table:

```dotenv
ACGW_DISCORD_ROLE_MAPPINGS=1550848038216409149=ac-core.admin
```

Use this when you want the mapping under version control or when you prefer not
to write to the database. The internal role must still exist in
`role_permissions` for it to grant anything.

### Demo store catalog

The demo admin role also receives the `gw.store.*` permissions. Seed an example
catalog (three products) to exercise the store:

```bash
psql "$ACGW_DATABASE_URL" -f scripts/seed_demo_store.sql
```

## Part 6 — Test the flow

1. Open a browser at
   `https://<tunnel-host>/api/v1/auth/discord/login` (or
   `http://localhost:8080/api/v1/auth/discord/login`).
   With `task dev` you can instead open `http://localhost:5173/` and click
   **Login with Discord** (Vite proxies `/api` to the gateway).
2. You are redirected to Discord and asked to authorize the app. Approve.
3. Discord redirects to the callback; the gateway sets an `HttpOnly` cookie and
   redirects to `ACGW_DISCORD_POST_LOGIN_REDIRECT_URL` (`http://localhost:8080/`
   shows the test page with the principal).
4. Verify the principal:

   ```bash
   # with the browser cookie jar, or paste the cookie value
   curl -i -b "acgw_session=<token>" https://<tunnel-host>/api/v1/me
   ```

   Expected `200` with `{"user_id":"...","discord_id":"...","roles":["member"]}`.
5. Verify the database never stores the raw token:

   ```sql
   SELECT id, user_id, expires_at, roles FROM sessions;
   -- id is a 64-char SHA-256 hex digest, not the 64-char cookie value
   SELECT discord_id, username, global_name FROM discord_identities;
   SELECT id, discord_id, roles, roles_synced_at FROM community_users;
   ```
6. Verify logout:

   ```bash
   curl -i -X POST -b "acgw_session=<token>" https://<tunnel-host>/api/v1/auth/logout
   # 204; the sessions row is gone and /me now returns 401
   ```
7. Negative checks:
   - `GET /callback?code=x` with no `state` -> `400 bad_request`.
   - Replay the same `state` -> `400 bad_request` (single use).
   - Point `ACGW_DISCORD_CLIENT_SECRET` at a wrong value -> `502 bad_gateway`.
8. Role change: remove a mapped Discord role from your test account, log in
   again, and confirm `GET /me` no longer lists it.

## Rollback

- Disable the feature by clearing `ACGW_DISCORD_CLIENT_ID` (or
  `ACGW_DISCORD_CLIENT_SECRET`) and restarting; the auth endpoints stop
  redirecting.
- Revert the migration with `task db:rollback` only on a database with no
  production sessions (it drops `oauth_states`, `discord_role_mappings` and the
  `roles` columns).
- To un-expose the callback, stop the `cloudflared` service and delete the
  redirect URI from the Discord portal.

## Troubleshooting

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| `invalid_client` at token exchange | wrong client id/secret or old secret | re-copy from portal; secret is shown once |
| `invalid_grant` at token exchange | redirect URI mismatch or reused/expired code | redirect URI must match **exactly**; codes are single use |
| `400 bad_request` on callback | missing/replayed/expired `state` (>10 min) | retry login; do not bookmark the callback |
| Discord shows "Invalid OAuth2 redirect_uri" | tunnel URL changed or http used | re-register the exact HTTPS callback |
| `guilds.members.read` returns 404/403 | user not in the guild, or scope not granted | join the configured guild; re-consent to grant the scope |
| `roles: []` after login | no mapping rows, or role id mismatch | insert mappings; copy role ids in Developer Mode |
| `/me` returns 401 after login | `Secure` cookie over http, or cookie blocked | use HTTPS via the tunnel; set `ACGW_SESSION_COOKIE_SECURE` correctly |
| `429 too_many_requests` | auth rate limit exceeded | wait for `Retry-After`, or raise `ACGW_AUTH_RATE_LIMIT_*` |
| Startup fails with a config error | insecure production config | see the error: secure cookie / HTTPS redirect required when `ACGW_ENV=production` |
| `502 bad_gateway` on callback | Discord unreachable / 5xx | check egress and `ACGW_DISCORD_TIMEOUT` |
| Discord configured but no redirect | one of the three required vars empty | set all three; check startup log |

## Secret handling

- Never paste `ACGW_DISCORD_CLIENT_SECRET`, tokens or session ids into tickets,
  logs or chat.
- Rotate credentials with
  [rotate-discord-secret.md](rotate-discord-secret.md); a secret rotation does
  not invalidate existing sessions.
- The tunnel token is a secret too; keep `CLOUDFLARE_TUNNEL_TOKEN` out of git.

## See also

- [discord-auth plan](../plan/discord-auth/README.md)
- [authentication.md](../authentication.md)
- [security.md](../security.md)
- [deploy.md](deploy.md), [migrations.md](migrations.md)
