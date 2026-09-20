# 013 — Metrics endpoint protection

**Phase:** 3 · **Gate:** G-standard, contributes to G3-hardening · **Depends on:** 012

## Goal

Ensure the observability endpoint is never exposed unauthenticated by accident,
and that the documented warning actually exists.

## Findings addressed

- Medium: `ACGW_METRICS_ENABLED=true` without `ACGW_METRICS_TOKEN` serves
  `/metrics` openly; config/validate never checks the combination; the hardening
  ticket claims a startup warning that does not exist.
- Low: `web/Caddyfile` reverse-proxies `/metrics` publicly (addressed with the
  SPA edge in ticket 039, but the server-side default is fixed here).

## Context

`internal/core/httpapi/metrics.go`, `server.go`, `config.go`, and
`docs/plan/discord-auth/ticket/005-hardening.md`. The metrics bearer check uses
`subtle.ConstantTimeCompare` and is correct when a token is set.

## Atomic change

Require a token when metrics are enabled outside development, and log a loud
warning otherwise. Optionally gate exposure behind an explicit allow flag.

## Requirements

- In non-development environments, refuse startup when
  `ACGW_METRICS_ENABLED=true` and `ACGW_METRICS_TOKEN` is empty.
- In development, allow but log a clear warning at startup.
- Keep the constant-time bearer comparison; add a `404` (not `401`) response when
  metrics are disabled so the endpoint is indistinguishable from an unknown path.
- Document the exact behaviour in `.env.example`, `docs/security.md` and
  `docs/authentication.md`.
- Do not log the token.

## Tests

- Test: enabled + no token + production config → startup error.
- Test: enabled + no token + development → warning logged, endpoint reachable.
- Test: enabled + token → `401` without `Authorization`, `200` with the correct
  bearer, `401` with a wrong bearer.
- Test: disabled → `404`.

## Acceptance criteria (gate G-standard, G3-hardening)

- `task check` green.
- It is impossible to expose metrics unauthenticated in production by
  configuration alone.

## Rollback

Revert; the open-by-default behaviour returns.

## Out of scope

- Edge/Caddy routing (ticket 039).
- Adding new metric families.
