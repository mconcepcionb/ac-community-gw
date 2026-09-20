# 005 — Authentication hardening

## Goal

Harden the authentication surface: rate-limit the OAuth endpoints, make cookie
attributes explicit and safe, fail fast on unsafe production configuration, and
expose counters for observability.

## Context

Tickets 001–004 deliver a working flow. This ticket closes the operational gaps
before the feature is exposed publicly.

## Scope

1. **Rate limiting** on `/api/v1/auth/discord/login` and `/callback`.
2. **Configurable `SameSite`** cookie attribute.
3. **Production safety checks** at startup.
4. **Observability counters** exposed at `/metrics`.

Out of scope: distributed rate limiting (single-instance token bucket is
enough for now), full Prometheus instrumentation of every subsystem.

## 1. Rate limiting

- In-memory token bucket keyed by client IP.
- Config: `ACGW_AUTH_RATE_LIMIT_PER_MINUTE` (default `30`),
  `ACGW_AUTH_RATE_LIMIT_BURST` (default `10`).
- Rejected requests return the shared JSON envelope with `429 too_many_requests`
  and a `Retry-After` header.
- Exposed to plugins as `Registry.RateLimit`, mirroring `RequireAuth`, so
  `identity-discord` wraps its two redirect endpoints without importing HTTP
  internals.
- Client IP: first `X-Forwarded-For` entry (the gateway runs behind
  cloudflared), else `RemoteAddr`.
- Idle buckets are swept so the map cannot grow unbounded.

## 2. Cookie attributes

- Add `ACGW_SESSION_COOKIE_SAMESITE` (`lax` default, `strict`, `none`).
- `SameSite=None` requires `Secure`.
- `auth.Manager` takes a `SameSite http.SameSite` option and uses it for
  `SetCookie` and `ClearCookie`.

## 3. Production safety checks

`config.validate` fails startup when `ACGW_ENV=production` and:

- `ACGW_SESSION_COOKIE_SECURE` is false;
- `ACGW_DISCORD_REDIRECT_URL` is set but not `https`.

It warns (does not fail) when `ACGW_METRICS_ENABLED=true` and no
`ACGW_METRICS_TOKEN` is set.

## 4. Observability

- New `internal/core/metrics` registry: named `int64` counters, a snapshot, and
  a Prometheus text handler.
- Config: `ACGW_METRICS_ENABLED` (default `false`), `ACGW_METRICS_TOKEN`
  (optional bearer token).
- `GET /metrics` is registered only when enabled; when a token is configured it
  requires `Authorization: Bearer <token>`.
- Counters added by `identity-discord`:
  - `identity_login_started_total`
  - `identity_login_completed_total`
  - `identity_login_failed_total`
  - `identity_session_created_total`
  - `identity_roles_changed_total`

## Acceptance criteria

- Exceeding the rate limit returns `429` with `Retry-After`; the limiter
  recovers after the refill window.
- `SameSite=None` without `Secure` is rejected by config validation.
- `ACGW_ENV=production` with an insecure cookie fails startup with a clear error.
- `/metrics` is absent when disabled, `401` without the token when configured,
  and returns counters when enabled.
- Existing tests and architecture boundaries keep passing.

## Tests

- Unit: token bucket consumes burst then denials then refills (injected clock);
  `ParseSameSite`; metric registry snapshot; metrics handler token check.
- Unit: plugin increments the login counters on success and failure.
- Config: production validation cases.
