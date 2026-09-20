# 006 — Trusted-proxy aware rate limiting

**Phase:** 2 · **Gate:** G-standard, contributes to G2-correctness · **Depends on:** 001

## Goal

Make the auth rate limiter use a trustworthy client identity so it cannot be
bypassed by spoofing `X-Forwarded-For`, and bound its memory use.

## Findings addressed

- High: `internal/core/httpapi/ratelimit.go` trusts the first `X-Forwarded-For`
  value unconditionally, so any client can pick a fresh key per request and
  defeat the login/callback rate limit. Between the 10-minute sweeps an attacker
  can also grow the bucket map without bound.

## Context

The limiter protects `/api/v1/auth/discord/login` and `/callback`. In the shipped
stack Caddy/cloudflared sanitize XFF, but the gateway cannot assume that when run
directly. There is currently no trusted-proxy configuration.

## Atomic change

Add a `TrustedProxies` configuration (CIDR list) and compute the client IP as
the rightmost non-trusted hop of `X-Forwarded-For`, falling back to
`RemoteAddr`. Add a hard cap on the number of tracked buckets.

## Requirements

- Add `ACGW_TRUSTED_PROXIES` (comma-separated CIDRs; empty means trust no proxy,
  i.e. use `RemoteAddr`).
- `clientIP` walks the XFF chain right-to-left, skipping trusted proxies, and
  returns the first untrusted address; if all hops are trusted, return
  `RemoteAddr`.
- Validate that `RemoteAddr` is within a trusted CIDR before honoring XFF at all.
- Cap `len(buckets)`; when at capacity, evict the oldest/idlest entries rather
  than growing (keep the periodic sweep as a secondary cleanup).
- Log at startup which proxies are trusted (counts/CIDRs, no per-request noise).
- Document the setting in `.env.example`, `docs/authentication.md` and
  `docs/security.md`.

## Tests

- Table tests for `clientIP`: no XFF (uses `RemoteAddr`), trusted peer + XFF,
  untrusted peer + spoofed XFF (ignored), multi-hop chain, malformed XFF.
- Replace the existing test that asserts unconditional XFF preference.
- Capacity test: feeding more unique keys than the cap keeps memory bounded and
  still rate-limits the most recent burst.
- Integration test: two requests with different spoofed XFF but the same
  `RemoteAddr` share one bucket when no proxy is trusted.

## Acceptance criteria (gate G-standard, G2-correctness)

- `task check` green; `go test -race ./...` green.
- A direct client cannot bypass the limit by changing XFF.
- Bucket count never exceeds the configured cap under load.

## Rollback

Revert to the previous `clientIP`. Expect the bypass to return; only revert if
the limiter is replaced.

## Out of scope

- Per-account rate limiting (separate feature).
- Changing the limit values.
