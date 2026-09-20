# Portal characters and self-mail

## Goal

Let players see their own characters, mail items or money to them, and choose
whether a character appears on public leaderboards.

## Context

A player must not use the staff-wide `GET /api/v1/azeroth/characters` list. Use
cases P4 and P5 ([../../../use-cases.md](../../../use-cases.md)) require
self-scoped access with ownership enforced server-side. The public surface (019)
exposes character names, so each character needs an opt-in flag.

## Requirements

- Backend: a self-scoped characters endpoint for the authenticated user
  (`GET /api/v1/azeroth/users/me/characters` or equivalent), resolved through the
  account link.
- Backend: a self-mail path that rejects a recipient that does not belong to the
  user's linked account.
- Backend: a per-character public-board opt-in flag the owner can toggle.
- Portal `/characters` lists only owned characters; `/characters/$name` shows
  detail, the self-mail form and the board opt-in toggle.
- Without a linked account, route to `/onboarding` instead of an empty state.

## Acceptance criteria

- A player sees only their characters; requesting another user's is forbidden.
- Self-mail to a non-owned character is rejected by the API.
- The opt-in toggle persists per character and is read by the leaderboards/public
  surface (018/019).
- New endpoints are annotated for OpenAPI and the client regenerated.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Ownership is enforced server-side; the SPA must never be the control.
- Reuse the character detail and mail form components from 006.
- Default the opt-in to off so no character is public by accident.

## Tests

- Go handler tests for ownership (own, other, unlinked) and the opt-in flag.
- RTL + MSW for the portal views and the toggle.

## Dependencies

- 001, 002. 011/012 provide onboarding for unlinked users; 018/019 read the
  opt-in.
