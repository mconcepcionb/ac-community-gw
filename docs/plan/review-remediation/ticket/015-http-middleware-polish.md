# 015 — HTTP middleware and error mapping polish

**Phase:** 3 · **Gate:** G-standard, contributes to G3-hardening · **Depends on:** 011

## Goal

Fix a set of core HTTP defects: panics missing from the access log, the
`statusRecorder` hiding optional interfaces, unvalidated request ids, an event-bus
panic, and inconsistent error/CORS-free defaults.

## Findings addressed

- Low: `recoveryMiddleware` sits outside `accessLogMiddleware`, so a panic is not
  access-logged.
- Low: `statusRecorder` implements only `WriteHeader`; it hides `Flush`, `Hijack`
  and `Unwrap`, breaking streaming/SSE and `http.ResponseController`.
- Low: client-supplied `X-Request-Id` is reflected and logged verbatim (log
  spoofing / correlation confusion).
- Low: `events.SubscribeTyped` panics for a pointer event type whose
  `EventName` has a value receiver.
- Low: `permissions.ErrNotRegistered` is dead code (never returned).
- Low: `Authorizer.RoleNames` documents "sorted" but does not sort.
- Low: `StatusForError` maps `azerothdb.ErrUnavailable`/`azerothcore.ErrNotConfigured`
  to a generic 500 instead of `503`/`502`.
- Low: `Server.New` defaults only `Metrics`; a nil `Readiness` panics and a nil
  `Audit` panics on record.
- Nit: `WriteJSON` ignores the encoder error.
- Nit: 401 responses omit `WWW-Authenticate`.

## Context

`internal/core/httpapi/*`, `internal/core/events/bus.go`,
`internal/core/permissions/*`. These are independent small fixes; grouping them
keeps the core HTTP layer consistent in one review.

## Atomic change

One pass over the HTTP layer and the small core defects above.

## Requirements

- Reorder middleware so access logging wraps recovery (still inside request-id
  injection) or log inside the recovery path.
- Add `Unwrap() http.ResponseWriter` to `statusRecorder` (and pass through
  `Flush`/`Hijack` if needed by any handler).
- Validate `X-Request-Id` against a strict charset/length; otherwise generate a
  fresh id.
- Make `SubscribeTyped` handle pointer event types safely (reflect on the type or
  document value-only events) and never panic at startup.
- Remove `ErrNotRegistered` or make a check path return it.
- Sort `RoleNames` (or fix the comment).
- Extend `StatusForError` with `azerothdb.ErrUnavailable` → 503 and
  `azerothcore.ErrNotConfigured` → 503/502.
- Default `Readiness` to an empty registry and `Audit` to `NopRecorder` in `New`.
- Return/handle the `WriteJSON` encoder error (log at least).
- Add `WWW-Authenticate` on 401.

## Tests

- Test: a panicking handler still produces an access-log entry.
- Test: `statusRecorder` exposes `Flush`/`Unwrap`.
- Test: an invalid `X-Request-Id` is replaced; a valid one is preserved.
- Test: `SubscribeTyped[*T]` does not panic.
- Test: `StatusForError` mapping table.
- Test: `New` with nil Readiness/Audit does not panic.

## Acceptance criteria (gate G-standard, G3-hardening)

- `task check` green; `go test -race ./...` green.
- No panic path remains in middleware construction.

## Rollback

Revert individually; the changes are independent and low risk.

## Out of scope

- Body limits (ticket 011).
- Config parsing (ticket 012).
