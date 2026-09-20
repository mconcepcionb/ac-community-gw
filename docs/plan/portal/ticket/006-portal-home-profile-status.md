# Portal home, profile and status

## Goal

Give players a personal dashboard, profile and server status under the portal
shell.

## Context

Today the home is a static card grid (`web/src/features/home/home-page.tsx`) and
`/profile` and `/azeroth/status` are flat routes. This ticket covers use cases
P9 and P11 and the community-overview intent of C5
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- `/` becomes the portal dashboard: linked-account status, points balance,
  characters summary, recent orders, recent announcements and a server status
  card.
- `/profile` moves into the portal (P11): identity, roles and permissions.
- `/status` moves into the portal (P9), authenticated for now; the public
  variant arrives in 016.
- Dashboard cards link into the portal use cases.
- A user with console permissions also sees the console link (001).

## Acceptance criteria

- The dashboard composes existing endpoints and degrades per card when a source
  is unavailable.
- A user without a linked account sees a prompt that routes to onboarding
  (010/011) instead of an empty state.
- Profile shows identity, roles and permissions; status shows parsed counts.
- `task web:check` green.

## Implementation notes

- Keep `/admin` (console overview) separate: the portal dashboard is personal,
  the console overview is operational and lands with the console tickets.

## Tests

- MSW for each card, including the unlinked-account path.

## Dependencies

- 001.
