# Console delivery

## Goal

Let staff search characters and deliver items or money in-game from the console.

## Context

Today `/characters` and its mail form serve both staff and (intended) players.
This ticket covers use case G4 and the staff half of P4/P5
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- `/admin/characters`: character search and detail for staff.
- `/admin/characters/$name`: mail delivery of items and/or money, with sanitised
  subject and body and the command result shown.
- Gate with `azeroth.character.list` and `azeroth.mail.send` for staff.
- Remove the flat `/characters` route here; the portal character views arrive in
  007.

## Acceptance criteria

- Staff can search any character and deliver to it; failures are retryable.
- The delivery result is shown and audited by the backend.
- The flat `/characters` route no longer exists.
- `task web:check` green.

## Implementation notes

- Reuse the existing mail form component; only rehome it under the console.
- Recipient validation stays server-side (letters only); the SPA mirrors it.

## Tests

- RTL + MSW for delivery success, validation failure and upstream failure.

## Dependencies

- 001.
