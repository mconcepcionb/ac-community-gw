# 039 — Caddy security headers

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 013

## Goal

Add browser hardening headers at the edge and stop proxying the metrics endpoint
publicly.

## Findings addressed

- Medium: `web/Caddyfile` sets only `X-Content-Type-Options` and `Referrer-Policy`;
  there is no `Content-Security-Policy`, `X-Frame-Options`,
  `frame-ancestors` or `Strict-Transport-Security`.
- Low: `/metrics` is reverse-proxied at the edge, so its protection depends
  entirely on the app-level token.

## Context

`web/Caddyfile`; the app mounts `/metrics` only when enabled (ticket 013).

## Atomic change

Add a restrictive header policy and remove the metrics proxy route (or restrict
it to an internal listener).

## Requirements

- Add a CSP that works with the built SPA (at minimum `default-src 'self'`,
  `object-src 'none'`, `base-uri 'self'`, `frame-ancestors 'none'`, and the
  necessary `script-src`/`style-src`/`connect-src` for the bundle; avoid
  `unsafe-inline` if possible, otherwise document a nonce/hash plan).
- Add `X-Frame-Options: DENY`, `Strict-Transport-Security` (with a sensible
  max-age), and keep `nosniff`/`Referrer-Policy`.
- Remove the public `/metrics` proxy route, or bind it to an internal-only
  handler; document how operators scrape metrics.
- Update `docs/frontend.md` / `docs/security.md` with the header policy.

## Tests

- Playwright smoke loads the SPA and asserts no CSP violations in the console and
  a successful login.
- `curl -I` asserts the headers are present.
- A request to `/metrics` from the edge returns `404` (or is unreachable) unless
  intended.

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green; Playwright smoke passes with no CSP violations.
- Required headers are present on the SPA origin.
- Metrics are not publicly proxied.

## Rollback

Revert the Caddyfile; headers disappear.

## Out of scope

- Application-level CSP nonces (only if required by the chosen policy).
