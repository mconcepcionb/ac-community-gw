# Public no-login surface

## Goal

Expose server status and leaderboards without authentication.

## Context

This ticket covers use cases O1 and O2 ([../../../use-cases.md](../../../use-cases.md)).
Public traffic needs its own rate limit.

## Requirements

- Backend: unauthenticated status and public leaderboard projections with a
  separate rate limit.
- Public payloads contain display data only, never account identifiers or
  personal data.
- Public views render for anonymous users on `/status` and `/leaderboards`.
- Authenticated users see the same data with their own highlighting.

## Acceptance criteria

- Anonymous status and boards render without a session.
- No account or personal data is exposed.
- Public traffic is rate-limited independently of authenticated traffic.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Reuse the status (006) and leaderboard (015) components; gate the personal
  highlighting on the session.
- The public routes must not require or redirect to login.

## Tests

- Go tests for the anonymous projection and the rate limit; RTL + MSW for the
  public views.

## Dependencies

- 006, 015.
