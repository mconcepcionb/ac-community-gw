# Public surface ownership

## Goal

Stop game plugins from serving game-specific data on the generic `/api/v1/public/*`
surface.

## Context

- `GET /api/v1/public/leaderboards/{board}` (`internal/plugins/azerothcharacter/plugin.go:111`).
- `GET /api/v1/public/status` (`internal/plugins/azerothinfo/plugin.go:69`).

`/public/*` is a gateway surface; the data is AzerothCore-specific. The routes
are unauthenticated and rate-limited/cached, and that behaviour must not change.

## Requirements

- Move the routes under the game prefix:
  - `GET /api/v1/azeroth/public/leaderboards/{board}`.
  - `GET /api/v1/azeroth/public/status`.
- Keep the same handlers, cache TTL and per-IP rate limit.
- Update the SPA public pages/hooks and the generated client.
- Keep `@ID`s stable where possible; update the paths only.

## Acceptance criteria

- No game-specific route remains under `/api/v1/public/*`.
- The public leaderboards and status pages work against the new paths.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- Option 2 from the plan README (a gateway `public` plugin that aggregates games)
  is deliberately deferred; it needs a multi-game public model. Moving the routes
  under the game prefix is the mechanical first step and keeps the no-login
  surface working.
- Document the deferred aggregator in the plan README status.

## Tests

- Update the Go route tests and the SPA tests that mock the public URLs.

## Dependencies

- Independent. Relates to [../../ux-improvements/ticket/010-online-structured.md](../../ux-improvements/ticket/010-online-structured.md)
  only through the shared character reader.
