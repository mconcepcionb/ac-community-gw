# Shared test support and core packages

**Milestone:** B · **Gate:** G-standard · **Depends on:** 001

## Goal

Introduce a small shared test-support package and cover the core packages that
report 0% today.

## Context

There is no shared test harness; every plugin declares per-package fakes and
builds requests by hand with `httptest` (e.g.
`internal/plugins/store/handlers_test.go`, `.../reports/handlers_test.go`), and
the shared bits are copy-pasted. `internal/core/{audit,azerothdb,ttlcache,plugins}`
have no tests at all; `logging` and `itemview` are thin.

## Requirements

- Add `internal/testsupport` with helpers only (proposal):
  `NewRequest(method, path, body)`, `SetPathValue`, `DecodeJSON`,
  `WithPrincipal(userID)`, `AssertStatus`. It may import only stdlib and
  `internal/core/*`; it must **never** import a plugin. If the architecture test
  classifies unknown packages, update `internal/architecture` to allow it while
  still forbidding plugin imports.
- Tests:
  - `core/ttlcache`: hit, miss, expiry, `Set` overwrite, concurrency.
  - `core/plugins`: `Manager.Add`/`RegisterAll`/`Names`, duplicate and empty-name
    errors; `Registry` route/permission/command/service/event registration and
    duplicate rejection.
  - `core/audit`: `NopRecorder`, `MultiRecorder` fan-out, `LogRecorder` output,
    and that no secret-bearing field is emitted.
  - `core/azerothdb`: row→domain mapping with a fake DB, `Unavailable*`
    variants, error sentinels, `ValidLeaderboard`.
  - `core/logging`: `New` level/format/output selection plus `ParseLevel`.
  - `core/itemview`: mapping branches (quality, class, subclass, inventory,
    bonding, damage, stats, resistances, `mapOr`).

## Tests

- The new tests, using `internal/testsupport` where an HTTP seam exists.
- `TestSourceIsFormatted` and the architecture tests stay green.

## Acceptance criteria

- `core/audit`, `core/azerothdb`, `core/ttlcache`, `core/plugins`,
  `core/logging`, `core/itemview` each reach ≥ 70%.
- `task check` green; `internal/testsupport` imports no plugin.

## Out of scope

- Plugin tests (C4-C8). C3 only establishes the harness and the core targets.
