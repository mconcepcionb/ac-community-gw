# 017 — Session cookie and store hardening

**Phase:** 3 · **Gate:** G-standard, contributes to G3-hardening · **Depends on:** 009

## Goal

Fix two session-layer defects: `Max-Age` computed from the wall clock instead of
the manager clock, and expired sessions that fail to delete being retried
forever.

## Findings addressed

- Low: `SetCookie` computes `MaxAge` with `time.Until(expires)` while the manager
  supports an injected `Now`; tests and clock-skewed hosts can compute `Max-Age`
  of 0 (session cookie) inconsistently.
- Nit: `Manager.Resolve` discards the delete error for an expired session, so a
  failed delete leaves expired rows forever.

## Context

`internal/core/auth/session.go`, `memory.go`, and the Postgres session store.
Login mints a fresh token (fixation defense) and only the SHA-256 hash is stored.

## Atomic change

Compute the cookie age from the manager clock and make expired-session deletion
observable and retried.

## Requirements

- Derive `Max-Age` from the manager's `now()` (or pass an explicit max age to
  `SetCookie`), so injected clocks behave correctly.
- On `Resolve` of an expired session, log/return the delete error (or at least
  surface it for metrics) instead of discarding it; keep returning
  not-found/anonymous to the caller.
- Ensure the cleanup job also removes expired rows (verify it covers the Postgres
  store, not only memory).
- Keep cookie flags (`HttpOnly`, `Secure`, `SameSite`) unchanged.

## Tests

- Test: with a frozen clock, the `Max-Age` matches `TTL` exactly.
- Test: an expired session is reported anonymous even when delete fails, and the
  failure is observable.
- Test: cleanup removes expired rows from the Postgres store.

## Acceptance criteria (gate G-standard, G3-hardening)

- `task check` green.
- Cookie age is clock-independent.

## Rollback

Revert; behaviour is low risk.

## Out of scope

- Role freshness (ticket 010).
- OAuth states (ticket 009).
