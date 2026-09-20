# 014 — Readiness redaction and timeouts

**Phase:** 3 · **Gate:** G-standard, contributes to G3-hardening · **Depends on:** 001

## Goal

Stop `/readyz` from leaking internal error details and from hanging on a stuck
dependency.

## Findings addressed

- Medium: `handleReadyz` embeds `result.Err.Error()` in the unauthenticated
  response, disclosing driver/connection details (host, port, database, user).
- Low: readiness checks run sequentially with only the request context; a stuck
  dependency hangs `/readyz` until the server/client timeout.

## Context

`internal/core/persistence/readiness.go` and `internal/core/httpapi/server.go`.
`postgres.Unavailable` carries the raw error via `Check`.

## Atomic change

Return coarse statuses in the response, log detailed errors server-side with the
request id, and bound each check with a per-check timeout.

## Requirements

- `/readyz` body reports only `ok`/`unavailable` (and the check name); never the
  underlying error string.
- Log the detailed error at `Warn`/`Error` with the request id and check name.
- Add a per-check timeout (configurable, small default) using
  `context.WithTimeout`; a timed-out check reports `unavailable`.
- Keep `/healthz` unchanged (liveness must not depend on dependencies).
- Apply the same redaction anywhere else an internal error is serialized
  (`StatusForError` already defaults to a generic 500; verify).

## Tests

- Test: an unavailable check returns `503` with no error text in the body.
- Test: the detailed error is emitted to the logger.
- Test: a blocking check times out and yields `unavailable` within the bound.

## Acceptance criteria (gate G-standard, G3-hardening)

- `task check` green.
- `grep` of the readiness response for a sentinel error string returns nothing.

## Rollback

Revert; the leak and hang return.

## Out of scope

- Changing which dependencies participate in readiness.
