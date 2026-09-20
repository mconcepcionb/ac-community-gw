# 004 — Role synchronization

## Goal

Map Discord guild roles to internal roles, apply the mapping on login, store the
effective roles in the session and load role -> permission grants into the
authorizer.

## Context

Discord roles are an external signal. Authorization stays local: the gateway
decides what a Discord role means. Discord role ids are numeric and are mapped
through a table owned by `identity-discord`; an unmapped role grants nothing.

## Requirements

- Fetch the user's guild member roles during login (`guilds.members.read`).
- Apply a DB-driven mapping (`discord_role_id` -> internal role).
- Store granted internal roles in the session/principal.
- Emit `identity.discord_roles_changed` when the effective roles change.
- Never grant permissions directly from an unmapped Discord role.
- No `IsAdmin`-style shortcut.
- Only mapped roles influence permissions.

## Deliverables

### 1. Mapping table + queries

Table in `migrations/00004_discord_oauth.sql` (see plan README). Queries:

```sql
-- name: ListInternalRolesForDiscordRoles :many
SELECT DISTINCT role
FROM discord_role_mappings
WHERE discord_role_id = ANY($1::text[]);

-- name: UpsertDiscordRoleMapping :exec
INSERT INTO discord_role_mappings (discord_role_id, role)
VALUES ($1, $2)
ON CONFLICT (discord_role_id) DO UPDATE
SET role = EXCLUDED.role, updated_at = now();

-- name: UpdateCommunityUserRoles :exec
UPDATE community_users
SET roles = $2, roles_synced_at = now(), updated_at = now()
WHERE id = $1;
```

`ListInternalRolesForDiscordRoles` takes a `[]string`; sqlc maps it to
`text[]`. Deduplicate and sort the result before it becomes the principal's
roles.

### 2. Resolver — `internal/plugins/identitydiscord/roles.go`

```go
func (p *Plugin) resolveRoles(
	ctx context.Context,
	accessToken, guildID string,
	user auth.DiscordUser,
) (current, changed []string, err error)
```

1. If `guildID == ""`, return the existing roles unchanged and no change.
2. `ids, err := p.provider.GuildMemberRoleIDs(ctx, accessToken, guildID)`.
3. `mapped, err := p.repo.ListInternalRolesForDiscordRoles(ctx, ids)`.
4. Read `community_users.roles` as the previous snapshot.
5. `changed` = set difference between `mapped` and previous (both directions).
6. `UpdateCommunityUserRoles(userID, mapped)`.
7. Return `mapped` as the principal roles.

Only the intersection with `discord_role_mappings` reaches the principal.
Discord's `@everyone` role id (equal to the guild id) is not special-cased and
grants nothing unless explicitly mapped.

### 3. Event

After the DB update commits, if `changed` is non-empty publish
`DiscordRolesChanged{UserID, DiscordID, Roles: mapped}`. On the first login
(previous empty, mapped non-empty) this is also emitted, so consumers see the
initial grant. No token is included.

### 4. Authorizer loading — `internal/core/permissions` + startup

The in-memory `Authorizer` starts empty. Load grants from `role_permissions` at
startup:

```sql
-- name: ListRolePermissions :many
SELECT role, permission FROM role_permissions;
```

Add a loader in `cmd/server` (or a small core helper) that calls
`authorizer.Grant(role, permissions...)` for each row after migrations are
applied. Log the number of roles loaded. If the database is absent, the
authorizer stays empty (deny all) — the local-dev login still works but no
permission-gated route is reachable.

Seeding `role_permissions` is operational (see the runbook); the gateway never
auto-grants permissions to a role from the Discord side.

### 5. Session integration

`handleCallback` passes the resolved roles into the principal used by
`auth.Manager.Create`, so `sessions.roles` and `principal.Roles` are the mapped
roles. `RequirePermission` then evaluates the local authorizer against them.
Roles are a login-time snapshot; changes take effect on the next login (ticket
004 acceptance note), matching the ticket's "next session" semantics.

## Acceptance criteria

- Mapping is configuration/DB-driven, not hardcoded.
- A mapped Discord role grants exactly its mapped internal role's permissions.
- An unmapped Discord role grants nothing.
- Removing a mapping (or a Discord role) revokes the permission on the next
  login.
- No permission check uses an `IsAdmin`-style shortcut.
- `DiscordRolesChanged` fires only when the effective roles change.

## Tests

- Unit: fake provider + fake repository; mapped role -> expected roles; unmapped
  -> empty; previous-equals-current -> no event; previous differs -> event.
- Integration: seed `discord_role_mappings` and `role_permissions`, run a login,
  assert the session roles and a `RequirePermission` check passes for a mapped
  role and fails for an unmapped one.
