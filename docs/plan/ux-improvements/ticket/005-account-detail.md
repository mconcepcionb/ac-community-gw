# Account detail page and claim status

## Goal

Give staff a single page that aggregates everything known about an account, and
surface whether the account is claimed by a community user.

## Context

- No `GET /api/v1/azeroth/accounts/{username}` route exists; account lookups are
  internal (`internal/core/azerothdb/accounts.go:46`,
  `internal/adapters/azerothmysql/mysql.go:98`).
- The account list DTO has no link/claim field
  (`internal/plugins/azerothaccount/accounts.go:20-31`).
- Ownership lives in `azeroth_account_links` (`migrations/00003_azeroth_account.sql:3-9`)
  and `account_claims` (`migrations/00010_account_claims.sql:3-11`); the only
  reverse lookup today is `GET /api/v1/admin/users/{id}` keyed by user, not
  account.
- Bans/ban reasons read from AzerothCore MySQL
  (`internal/adapters/azerothmysql/mysql.go:29-34`); staff actions are in the
  audit log.

## Requirements

- Backend `GET /api/v1/azeroth/accounts/{username}` returning the account plus:
  - `link`: `{ user_id, display_name, discord_id, linked_at }` when claimed.
  - `claim`: pending claim status when applicable, otherwise `unclaimed`.
  - `characters` summary (name/level/class/online) for the account.
  - `bans` history/reason.
- Extend `GET /api/v1/azeroth/accounts` (list) with `claimed` and `claimed_by`
  so the column can be rendered without an N+1.
- Claim status vocabulary: `unclaimed | claimed | pending_claim`.
- Console route `/admin/accounts/$username` with:
  - identity/status/moderation header using the shared controls (006),
  - linked community-user card linking to `/admin/users/$id`,
  - characters table linking to `/admin/characters/$name`,
  - **Annotations** panel (002),
  - **History** panel from audit (003).
- Permission: reuse the account read/list permission; mutations keep their
  existing permissions.

## Acceptance criteria

- The list shows a claim column; an unclaimed account and a claimed one render
  correctly, with a link to the owning user.
- The detail page loads by username and renders link, claim, characters, bans,
  annotations and history.
- `username` is URL-encoded; a missing account yields a 404 state, not a crash.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- The list join is Postgres-side (`azeroth_account_links`) enriched onto the
  MySQL read; do not fan out per row.
- Reuse `handleListCharacters` with the resolved `account_id` for the character
  summary rather than a new query.

## Tests

- Go handler tests for found/not-found and claimed/unclaimed.
- RTL + MSW tests for the detail page and the new list column.

## Dependencies

- 001 (primitives), 002 (annotations), 003 (history), 004 (user read link).
