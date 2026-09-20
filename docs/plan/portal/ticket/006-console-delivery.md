# Console delivery

## Goal

Let staff search characters and deliver items or money in-game from the console.

## Context

Today `/characters` and its mail form serve both audiences. This ticket covers
use cases G3 and G4 and the staff half of P4/P5
([../../../use-cases.md](../../../use-cases.md)). Character ban/unban (G3) was
deferred here from ticket 003 so it lives on the character detail.

## Requirements

- `/admin/characters`: character search and detail for staff.
- `/admin/characters/$name`: mail delivery of items and/or money, with sanitised
  subject and body and the command result shown, plus character ban/unban.
- Gate with `azeroth.character.list`, `azeroth.mail.send` and
  `azeroth.admin.characters.ban` for staff.
- Remove the flat `/characters` route here; the portal character views arrive in
  009.

## Acceptance criteria

- Staff can search any character, deliver to it and ban/unban it; failures are
  retryable.
- The delivery and ban results are shown and audited by the backend.
- The flat `/characters` route no longer exists.
- `task web:check` green.

## Implementation notes

- Reuse the existing mail form component; only rehome it under the console.
- Recipient validation stays server-side (letters only); the SPA mirrors it.

## Tests

- RTL + MSW for delivery success, validation failure and upstream failure.

## Dependencies

- 001, 002.
