# 002 — Session persistence

## Goal

Persist sessions in PostgreSQL, store only a hash of the session token, enforce
expiry and clean up expired rows.

## Context

The foundation ships `auth.Manager` + `auth.MemoryStore` (`internal/core/auth`).
ADR 0004 commits to PostgreSQL-backed server-side sessions. The generated
queries for `sessions` already exist but are not wired, and the `sessions`
table stores the token verbatim in `id`.

## Requirements

- Implement `auth.Store` using the generated `sessions` queries.
- Store only a hash (SHA-256 hex) of the session token, never the raw token.
- Persist the session's effective internal roles.
- Enforce expiry on resolve and delete expired rows.
- Add a periodic janitor for expired sessions and OAuth states.
- Wire the store in `cmd/server`; keep the memory store for tests.

## Deliverables

### 1. Token hashing in `auth.Manager`

Add an optional hasher, defaulting to SHA-256 hex, and hash at the single
source of truth (the manager), so every store stays dumb:

```go
type ManagerOptions struct {
	Store      Store
	TTL        time.Duration
	Secure     bool
	CookieName string
	Now        func() time.Time
	HashToken  func(string) string // default: sha256 hex
}
```

- `Create`: `token := newToken(); session.ID = m.hash(token)`; return the raw
  `token` to the caller for the cookie.
- `Resolve`: `m.store.Get(ctx, m.hash(token))`.
- `SetCookie`/`ClearCookie`/`TokenFromRequest` are unchanged (raw token in the
  cookie, hash in the database).

Update the `Session` doc comment to state that `ID` is the **hash** of the
cookie value.

### 2. Schema — part of `migrations/00004_discord_oauth.sql`

```sql
ALTER TABLE sessions ADD COLUMN roles text[] NOT NULL DEFAULT '{}';
-- community_users, discord_role_mappings, oauth_states in tickets 003/004/001
```

`discord_id` is recovered with a join to `community_users`; it is not stored on
`sessions`.

### 3. Queries — `internal/plugins/identitydiscord/repository/queries.sql`

Replace the session queries so they carry roles and the Discord id:

```sql
-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at, roles)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSession :one
SELECT s.id, s.user_id, s.created_at, s.expires_at, s.roles, cu.discord_id
FROM sessions s
JOIN community_users cu ON cu.id = s.user_id
WHERE s.id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < now();
```

Regenerate with `task codegen:sqlc`; update `generated/models.go` consumers
(`Session` gains `Roles []string`; `GetSession` gains `DiscordID`).

### 4. Store — `internal/plugins/identitydiscord/repository/store.go`

Package `repository`, importing the generated package and `internal/core/auth`.

```go
type SessionStore struct{ q *identitydiscordrepo.Queries }

var _ auth.Store = (*SessionStore)(nil)

func (s *SessionStore) Create(ctx context.Context, session auth.Session) error {
	return s.q.CreateSession(ctx, identitydiscordrepo.CreateSessionParams{
		ID:        session.ID,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
		Roles:     session.Roles,
	})
}

func (s *SessionStore) Get(ctx context.Context, id string) (auth.Session, error) {
	row, err := s.q.GetSession(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	if err != nil {
		return auth.Session{}, err
	}
	return auth.Session{
		ID:        row.ID,
		UserID:    row.UserID,
		DiscordID: row.DiscordID,
		Roles:     row.Roles,
		CreatedAt: row.CreatedAt,
		ExpiresAt: row.ExpiresAt,
	}, nil
}

func (s *SessionStore) Delete(ctx context.Context, id string) error {
	return s.q.DeleteSession(ctx, id)
}
```

### 5. Janitor — `internal/plugins/identitydiscord/janitor.go`

A `Run(ctx)` loop (started with `go` from `cmd/server`, stopped by the root
context) that every `ACGW_SESSION_CLEANUP_INTERVAL` (default `15m`) calls
`DeleteExpiredSessions` and `DeleteExpiredOAuthStates`, logging errors but not
exiting. Add `ACGW_SESSION_CLEANUP_INTERVAL` to config.

### 6. Wiring — `cmd/server/main.go`

```go
var sessionStore auth.Store = auth.NewMemoryStore()
var repo *identitydiscordrepo.Queries
if database != nil {
	repo = identitydiscordrepo.New(database.SQL())
	sessionStore = repository.NewSessionStore(repo)
}
sessions := auth.NewManager(auth.ManagerOptions{Store: sessionStore, /* ... */})
```

When no database is configured the gateway still starts on the memory store,
matching today's behaviour; log a warning.

## Acceptance criteria

- A raw session token never appears in the `sessions` table.
- Sessions survive a process restart (PostgreSQL configured).
- Expired sessions are rejected and deleted on resolve; the janitor removes
  expired rows and OAuth states.
- Logout revokes the persisted session.
- `DELETE FROM sessions` / a database dump reveals no usable token.

## Tests

- Unit: `auth.Manager` hashes before `Create` and before `Get` (fake store
  asserts the key), using a deterministic hasher.
- Integration (`//go:build integration`): create -> restart connection ->
  resolve; expiry; `DeleteExpiredSessions`; logout.
