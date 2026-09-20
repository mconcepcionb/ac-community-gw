# Public no-login surface

## Goal

Expose server status and named character leaderboards without authentication,
with a separate rate limit and a short cache.

## Context

This ticket covers use cases O1 and O2
([../../../use-cases.md](../../../use-cases.md)). The status endpoint is gated by
`azeroth.info.public.read` today; the public surface makes the chosen data
anonymous.

## Requirements

- Backend: unauthenticated status (`connected`, `peak`, `queue`, `uptime`) and
  public leaderboard projections.
- Public leaderboards show **real character names**, but only for characters
  whose owner opted in (flag from 009).
- A short server-side cache plus a **separate** per-IP rate limit for public
  endpoints.
- Public payloads contain display data only: no account identifiers, no Discord
  data, no player email.
- Public views render for anonymous users on `/status` and `/leaderboards`;
  authenticated users see the same data with their own highlighting.

## Acceptance criteria

- Anonymous status and boards render without a session.
- Only opted-in characters appear on public boards; no non-display fields leak.
- Public traffic is cached and rate-limited independently of authenticated
  traffic.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Reuse the status (008) and leaderboard (018) components; gate personal
  highlighting on the session.
- The public routes must not require or redirect to login.

## Tests

- Go tests for the anonymous projection, the opt-in filter, caching and the rate
  limit; RTL + MSW for the public views.

## Dependencies

- 008, 009, 018.
