# Two-surface shell

## Goal

Introduce the portal and console shells with permission-based landing and
navigation, so every later ticket can relocate a feature into the right surface.

## Context

The SPA has a single shell and a flat nav (`web/src/app/app-nav.tsx`) listing
every feature. The use cases define a player **portal** at `/` and a staff
**console** at `/admin/*` (see [../../../use-cases.md](../../../use-cases.md),
sections *Surfaces* and *Information architecture*).

## Requirements

- Split the root route into two layouts: a portal shell at `/` and a console
  shell at `/admin`.
- Landing resolver: after sign-in, a user holding any console permission lands
  on `/admin`; everyone else lands on `/`.
- Navigation is permission-driven and surface-specific. The portal never renders
  console links.
- `/admin` renders a placeholder overview that 002-005 fill in.
- A link lets a user with both surfaces move between them.
- Existing feature pages stay reachable during the transition; they move in
  002-008 and are removed in 009.

## Acceptance criteria

- Anonymous users are sent to login; authenticated users land per the resolver.
- A portal-only user cannot reach `/admin/*`: the route resolves to a forbidden
  state with no console chrome.
- A console user can move between portal and console.
- `task web:check` green.

## Implementation notes

- Add layout components under `web/src/app/`; reuse `components/common` and
  `components/ui`.
- Add a `hasAnyConsolePermission` helper backed by a `CONSOLE_PERMISSIONS` list.
- No API or generated-client changes.

## Tests

- Route-level tests for the landing resolver (staff, player, anonymous) and the
  forbidden state.

## Dependencies

- None. Unblocks every other ticket.
