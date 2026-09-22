# One-command reset

## Goal

`task demo:reset` restores a clean slate between runs.

## Requirements

- `POST http://localhost:7878/reset` to clear the fake AzerothCore state.
- Re-run the idempotent seeds.
- Do nothing harmful when the fake server is not running.

## Acceptance criteria

- `task demo:reset` leaves the demo ready to repeat.

## Dependencies

- 001, 002.
