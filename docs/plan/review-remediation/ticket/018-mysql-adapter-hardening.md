# 018 — MySQL adapter hardening

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 001

## Goal

Make the AzerothCore MySQL adapters robust: force time parsing, bound
pagination, escape `LIKE`, use context-aware pings, and remove duplicated pool
setup.

## Findings addressed

- Medium: the adapter scans `last_login` into `sql.NullTime` but neither sets nor
  validates `parseTime=true`; a custom DSN fails at query time.
- Medium: `offset` is only floored at 0; deep offsets force a huge scan.
- Medium: `LIKE` wildcards in user filters are unescaped (`%`/`_`), so a filter of
  `%` scans the whole table.
- Low: `db.Ping()` without context can hang startup (Postgres already uses
  `PingContext`).
- Low: three near-identical pool constructors; no `SetConnMaxIdleTime`; DSN flags
  such as `multiStatements`/`interpolateParams` are unrestricted while the
  adapter is meant to be read-only.
- Low: `postgres.Open` applies zero pool values unconditionally (`SetMaxOpenConns(0)`
  = unlimited).
- Low: `goose.SetBaseFS`/`SetDialect` mutate package globals and race under
  parallel calls; `MigrateSQL` duplicates `Migrate`.

## Context

`internal/adapters/azerothmysql/*`, `internal/adapters/postgres/*`. Queries are
parameterized; these are robustness and operational fixes, not injection fixes.

## Atomic change

Introduce one shared `openPool` helper for the MySQL adapters and fix the
pagination/filter/ping/goose issues. PostgreSQL pool guards are included because
they share the same defect class.

## Requirements

- Parse the DSN with `mysql.ParseDSN`, force `ParseTime = true` (and a defined
  `Loc`), then `sql.Open(..., dsn.FormatDSN())`; reject unsupported flags
  (`multiStatements`, `interpolateParams`) for the read-only client.
- Clamp `offset` to a configured maximum (or cap `offset+limit`); return an error
  beyond it.
- Escape `\`, `%`, `_` in filter values and use `ESCAPE '\\'`.
- Replace `db.Ping()` with `PingContext` using a bounded timeout.
- Extract a shared `openPool(dsn, cfg)`; set `SetConnMaxIdleTime`; use a small
  max-open pool for the read-only adapters.
- Guard Postgres pool setters with `> 0` or supply defaults.
- Serialize goose usage (mutex or provider API) and delete the duplicate
  `MigrateSQL`.

## Tests

- Test: a DSN without `parseTime` is corrected (or rejected) so timestamp scans
  succeed.
- Test: an over-max offset is rejected; `limit` remains clamped.
- Test: a filter containing `%`/`_` does not act as a wildcard.
- Test: `PingContext` is used (cancelled context returns promptly).
- Integration tests against MySQL for the above (behind `-tags integration`).
- Postgres: zero-valued pool config does not produce an unbounded pool.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green; integration tests pass.
- No context-less network call remains in the adapters.
- One pool constructor is shared by the MySQL adapters.

## Rollback

Revert individual changes; the DSN hardening is the only behaviour-visible one.

## Out of scope

- Query result shapes.
- SQL injection (queries are already parameterized).
