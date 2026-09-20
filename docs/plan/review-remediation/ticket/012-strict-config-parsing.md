# 012 — Strict configuration parsing

**Phase:** 3 · **Gate:** G-standard, contributes to G3-hardening · **Depends on:** 001

## Goal

Make invalid environment values fail startup instead of silently falling back to
a default, matching the documented configuration contract.

## Findings addressed

- Medium: `envBool`, `envInt` and `envDuration` return the fallback on parse
  error, so `ACGW_SESSION_TTL=25hours`, `ACGW_METRICS_ENABLED=treu` or
  `ACGW_DISCORD_TIMEOUT=abc` are silently ignored and `validate` cannot catch
  them (the package doc promises a startup error).
- Nit: an empty variable is treated as unset, so a value cannot be explicitly
  configured to empty.
- Nit: `logging.ParseLevel` accepts `"warning"` while `config.validate` rejects
  it — inconsistent level names.

## Context

`internal/core/config/config.go`, `internal/core/logging/logging.go`. All values
are read from environment variables; `.env.example` is the documented template.

## Atomic change

Make the typed env helpers return `(value, error)`, accumulate errors in `Load`,
and fail with a single message listing every invalid variable. Align log-level
parsing across packages.

## Requirements

- `envBool`/`envInt`/`envDuration` return an error on parse failure; `Load`
  aggregates and returns them together (variable name + offending value, never a
  secret value).
- Decide and document the empty-string semantics; if empty remains "unset",
  document it, and add an explicit way to set an empty value if needed.
- Unify accepted log levels (`debug|info|warn|warning|error`) between `config`
  and `logging`, or reject `warning` in both.
- Keep defaults unchanged for absent variables.
- Add the request-body limit setting from ticket 011 here if it is not already
  configurable.
- Update `.env.example` with any new/changed keys.

## Tests

- Table tests: valid values parse; invalid values produce a startup error naming
  the variable; absent values use defaults.
- Test: multiple invalid values are all reported.
- Test: a secret value (e.g. a token) is never echoed in the error message.
- Test: log-level parsing accepts/rejects the same set in both packages.

## Acceptance criteria (gate G-standard, G3-hardening)

- `task check` green.
- `ACGW_SESSION_TTL=25hours` prevents startup with a clear error.
- Documentation matches behaviour.

## Rollback

Revert; invalid values silently fall back again.

## Out of scope

- Adding new settings beyond those required by tickets 011 and 013.
- Secret management (ticket 029).
