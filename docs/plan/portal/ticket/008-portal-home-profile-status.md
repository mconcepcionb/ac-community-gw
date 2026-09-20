# Portal home, profile and status

## Goal

Give players a personal dashboard, profile and server status under the portal
shell.

## Context

Today the home is a static card grid and `/profile` and `/azeroth/status` are
flat routes. This ticket covers use cases P9 and P11
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- `/` becomes the portal dashboard: linked-account status, points balance,
  characters summary, recent orders, recent announcements and a server status
  card.
- `/profile` moves into the portal (P11): identity, roles and permissions.
- `/status` moves into the portal (P9), authenticated for now; the public
  variant arrives in 019.
- Dashboard cards link into the portal use cases.
- A user with console permissions also sees the console link (002).

## Acceptance criteria

- The dashboard composes existing endpoints and degrades per card when a source
  is unavailable.
- A user without a linked account sees a prompt that routes to onboarding
  (011/012) instead of an empty state.
- Profile shows identity, roles and permissions; status shows parsed counts.
- `task web:check` green.

## Implementation notes

- Keep the console overview (007) separate: this dashboard is personal, that one
  is operational.

## Tests

- MSW for each card, including the unlinked-account path.

## Dependencies

- 001, 002.
