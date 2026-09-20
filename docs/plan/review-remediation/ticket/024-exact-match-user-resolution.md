# 024 — Exact-match user resolution

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 001

## Goal

Make directory lookups by name deterministic so account linking never misses a
real user because of pagination.

## Findings addressed

- Medium: `ResolveUserByName` calls `ListUsers(name, 20, 0)` (a fuzzy `%name%`
  search ordered by `created_at DESC`) and filters for exact matches in Go. When
  20+ users contain the substring, the exact match may not be fetched, producing a
  spurious `user_not_found` or breaking ambiguity detection during account
  linking.

## Context

`internal/plugins/identitydiscord/plugin.go` (the user directory capability) and
`repository/queries.sql`. The SQL already supports a filter; this ticket adds a
dedicated exact-match query.

## Atomic change

Add an exact-match query and use it for directory resolution; keep the fuzzy
query for the search endpoint.

## Requirements

- New query `ResolveUserByName`:
  `WHERE lower(cu.display_name) = lower($1) OR lower(di.username) = lower($1) OR
   lower(di.global_name) = lower($1) OR cu.discord_id = $1` (discord_id exact).
- Return **all** exact matches so the caller can detect ambiguity (multiple users
  with the same display name) and return a typed ambiguity error.
- The directory `ResolveUserByName` uses the exact query; the HTTP search endpoint
  keeps the fuzzy query.
- Escape `%`/`_` in the fuzzy query (also covered by ticket 021).
- Regenerate sqlc (ticket 028 cleans unused queries; this adds one).

## Tests

- Test: an exact match is found even when 20+ fuzzy matches exist.
- Test: case-insensitive matching works for display name/username/global name.
- Test: two users with the same name produce an ambiguity error, not a random
  pick.
- Integration test against Postgres.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green.
- `ResolveUserByName` never depends on a fuzzy page for correctness.

## Rollback

Revert; the fuzzy-based resolution returns.

## Out of scope

- Changing the search endpoint's ranking.
- The user table schema (indexes in ticket 021).
