# 009 — OAuth state expiry and store parity

**Phase:** 2 · **Gate:** G-standard, contributes to G2-correctness · **Depends on:** 001

## Goal

Make the persisted OAuth state store enforce expiry exactly like the in-memory
store, so stale states cannot be consumed.

## Findings addressed

- High: `ConsumeOAuthState` deletes and returns the row regardless of
  `expires_at`; the store discards `row.ExpiresAt`. The in-memory store enforces
  expiry, so behaviour differs by backend and stale states are accepted until the
  cleanup job runs, weakening OAuth CSRF/state protection.

## Context

`internal/plugins/identitydiscord/repository/queries.sql`, `store.go`, and
`memory.go`. States are single-use and short-lived; the callback consumes them.

## Atomic change

Add an expiry predicate to the consume query and assert parity between the
memory and Postgres stores with a shared test suite.

## Requirements

- `ConsumeOAuthState` becomes
  `DELETE FROM oauth_states WHERE state = $1 AND expires_at > now() RETURNING ...`
  (or check `row.ExpiresAt` in Go); an expired/absent state maps to
  `ErrStateNotFound`.
- Confirm the store never logs or returns the PKCE verifier on failure.
- Add a table-driven conformance test parameterized over both stores
  (`MemoryStateStore`, `repository.Store`) asserting identical behaviour for:
  unknown state, valid state, expired state, replayed state.
- Ensure the cleanup job and the consume path agree on the clock source
  (`now()` in SQL, injected `Now` in memory) and document it.

## Tests

- Conformance test as above; the expired-state case must fail before the fix.
- Integration test against Postgres for the DELETE ... AND expires_at predicate.
- Callback handler test: an expired state returns the same error as an unknown
  state.

## Acceptance criteria (gate G-standard, G2-correctness)

- `task check` green; integration tests pass.
- Both stores reject expired states identically.

## Rollback

Revert; the Postgres store reverts to accepting stale states.

## Out of scope

- Session expiry (ticket 017).
- OAuth flow hardening beyond state expiry (ticket 034 handles the SPA side).
