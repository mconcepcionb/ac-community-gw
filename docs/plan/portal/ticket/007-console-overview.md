# Console overview

## Goal

Give staff an operations overview that summarises the community and links into
the console use cases.

## Context

Use case C5 ([../../../use-cases.md](../../../use-cases.md)). Today the home is a
static card grid (`web/src/features/home/home-page.tsx`); the portal dashboard
(008) is personal, this one is operational.

## Requirements

- `/admin` shows: server status, queue counts (once 015 exists), recent
  announcements, recent moderation activity and recent store activity.
- Each card links into the relevant console use case.
- Figures respect the viewer's permissions: never show data the viewer cannot
  open.

## Acceptance criteria

- The overview composes existing endpoints and degrades per card when a source is
  unavailable.
- Cards the viewer cannot access are hidden.
- The placeholder from 001 is replaced.
- `task web:check` green.

## Implementation notes

- Reuse the status card from 008 and the stat primitives from `components/common`.
- Queue counts appear only after 015; hide the card until then.

## Tests

- MSW for each card, including partial failure and permission-hidden cards.

## Dependencies

- 003-006. Extended by 015 (queue counts).
