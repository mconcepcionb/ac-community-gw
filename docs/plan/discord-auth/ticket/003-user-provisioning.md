# 003 — User provisioning

## Goal

Create or update the community user for an authenticated Discord identity,
idempotently, and announce it on the event bus.

## Context

On first login the gateway must map a Discord identity to a stable
`community_users` row keyed by `discord_id`. `discord_identities` holds the
mutable profile. The generated `CreateCommunityUser` / `UpsertDiscordIdentity`
queries exist but are not wired and the create path is not idempotent under
concurrency.

## Requirements

- On first login insert `community_users` and `discord_identities`.
- On later logins update the Discord profile (username, global name, avatar).
- Identity key is `discord_id`, never the username.
- Concurrent first logins must not create duplicates.
- Emit `identity.user_provisioned` and `identity.user_authenticated`.
- Publish events only after the transaction commits; events carry identifiers
  only, never tokens.

## Deliverables

### 1. Idempotent queries — `queries.sql`

```sql
-- name: UpsertCommunityUser :one
INSERT INTO community_users (id, discord_id, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (discord_id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    updated_at = now()
RETURNING *, (xmax = 0) AS inserted;

-- name: UpsertDiscordIdentity :one
INSERT INTO discord_identities (user_id, discord_id, username, global_name, avatar)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE
SET username = EXCLUDED.username,
    global_name = EXCLUDED.global_name,
    avatar = EXCLUDED.avatar,
    updated_at = now()
RETURNING *;
```

`xmax = 0` in the `RETURNING` list is true only for the inserting statement,
which gives a race-free "was this created now" signal without a second query.
If reading it from sqlc is awkward, fall back to
`GetCommunityUserByDiscordID` first and accept a possible duplicate
`UserProvisioned` event under a race (the unique constraint still prevents a
duplicate row).

Display name fallback: `global_name`, else `username`, else `discord_id`.

### 2. Provisioner — `internal/plugins/identitydiscord/provision.go`

```go
type provisionResult struct {
	UserID   uuid.UUID
	Created  bool
	Identity identitydiscordrepo.DiscordIdentity
}

func (p *Plugin) provision(ctx context.Context, u auth.DiscordUser) (provisionResult, error)
```

Steps:

1. `tx, err := p.db.BeginTx(ctx, nil)`.
2. `q := p.repo.WithTx(tx)`.
3. `user := q.UpsertCommunityUser(...)` with a fresh `uuid.New()` and the
   display name.
4. `identity := q.UpsertDiscordIdentity(user.ID, u.ID, u.Username, u.GlobalName, u.Avatar)`.
5. `tx.Commit()`.
6. After commit, return the result; the caller publishes events.

No Discord token, access token or session token ever enters this function.

### 3. Events — caller in `oauth.go`

After commit (and after role sync, ticket 004, for the authenticated event):

- `provisioned == true` -> `reg.Events.Publish(ctx, UserProvisioned{UserID, DiscordID})`
- always -> `reg.Events.Publish(ctx, UserAuthenticated{UserID, DiscordID})`

`events.go` already defines the structs. Publish synchronously; a handler error
is logged but must not fail the login (the session is already created), or the
publish is done before the response and errors are surfaced as `500`. Choose one
and document it; the plan uses **log-and-continue** for `UserAuthenticated` and
**fail-fast** for `UserProvisioned` before the response is written.

### 4. Audit

Record via `reg.Audit` (or the injected recorder):

- `action="identity.login"`, `result=success|failure`, `target_type="community_user"`,
  `target_id=<user_id>`, `actor_discord_id=<discord_id>`, request id.
- Never include tokens or the `code`.

## Acceptance criteria

- Provisioning twice yields one `community_users` row and one
  `discord_identities` row.
- Two concurrent first logins yield exactly one `community_users` row.
- Changing the Discord username does not create a new `user_id`.
- `UserProvisioned` is emitted only when the user was created; both events carry
  identifiers only.

## Tests

- Unit with a fake repository: create-then-update path, display-name fallback,
  `Created` flag semantics, one `UserProvisioned` on second login.
- Integration (`//go:build integration`): parallel `provision` calls, assert a
  single row; simulate a username change and assert the same `user_id`.
