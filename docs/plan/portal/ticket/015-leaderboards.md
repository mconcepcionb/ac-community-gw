# Leaderboards

## Goal

Give players rankings for character progression, wealth, playtime and
PvP/arena.

## Context

This ticket covers use case P10 ([../../../use-cases.md](../../../use-cases.md))
and needs new read-only queries against the character DB.

## Requirements

- Backend: paginated board queries for character progression, wealth,
  playtime/activity and PvP/arena.
- Queries are cached and refreshed on an interval; each board is individually
  disableable.
- Portal `/leaderboards` and `/leaderboards/$board`: paginate and highlight the
  signed-in user's own position.
- Board definitions are permission-gated.

## Acceptance criteria

- Each board paginates and highlights the current user when present.
- An unavailable or disabled board is hidden gracefully.
- Queries are bounded and cached so the read-only database is not overloaded.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- The character schema already exposes `level`, `money`, `totaltime` and
  `arenaPoints`, so each board can be derived without schema changes.

## Tests

- Go tests for ordering, pagination and disable; RTL + MSW for the boards.

## Dependencies

- 001, 006. 016 exposes the public projection.
