# Leaderboards

## Goal

Give players rankings for character progression, wealth, playtime and PvP/arena.

## Context

This ticket covers use case P10 ([../../../use-cases.md](../../../use-cases.md))
and needs new read-only queries against the character DB. The character schema
already exposes `level`, `money`, `totaltime` and `arenaPoints`
(`scripts/mysql/init/02_acore_characters.sql`).

## Requirements

- Backend: paginated board queries for character progression, wealth,
  playtime/activity and PvP/arena.
- Queries are cached and refreshed on an interval; each board is individually
  disableable.
- Portal `/leaderboards` and `/leaderboards/$board`: paginate and highlight the
  signed-in user's own position.
- Boards include only characters whose owner has opted in (flag from 009).
- Board definitions are permission-gated.

## Acceptance criteria

- Each board paginates and highlights the current user when present.
- A character without the opt-in flag never appears on a board.
- An unavailable or disabled board is hidden gracefully.
- Queries are bounded and cached so the read-only database is not overloaded.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Derive each board from existing columns; no schema changes.
- The opt-in filter is applied server-side before ranking is exposed.

## Tests

- Go tests for ordering, pagination, opt-in filtering and disable; RTL + MSW for
  the boards.

## Dependencies

- 001, 002, 008. 009 provides the opt-in flag; 019 exposes the public projection.
