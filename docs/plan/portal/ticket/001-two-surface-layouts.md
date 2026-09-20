# Two-surface layouts

## Goal

Introduce the portal and console route layouts so later tickets can relocate
features into the correct surface.

## Context

The SPA has a single shell and a flat nav
(`web/src/app/app-nav.tsx`, `web/src/app/app-shell.test.tsx`). The use cases
define a player **portal** at `/` and a staff **console** at `/admin/*` (see
[../../../use-cases.md](../../../use-cases.md), section *Surfaces*).

## Requirements

- Add a portal layout rendered at `/` and a console layout rendered at `/admin`.
- Console routes are nested under `/admin/*`; the two layouts never share a
  parent route with a mixed nav.
- Each layout provides its own navigation slot; shared header and footer
  primitives stay common to both.
- `/admin` renders a placeholder index (completed by 007).
- Existing feature pages stay reachable during the transition; they move in
  003-010 and are removed in 013.
- No landing logic here (that is 002) and no API or generated-client changes.

## Acceptance criteria

- `/` renders the portal layout; a `/admin` route renders the console layout.
- Existing pages still resolve and render.
- `task web:check` green.

## Implementation notes

- Add layout components under `web/src/app/`; reuse `components/common` and
  `components/ui`.
- Keep the guard primitives from the auth feature; the forbidden state lands in
  002.

## Tests

- Route-level render tests for both layouts.

## Dependencies

- None. Unblocks every other ticket.
