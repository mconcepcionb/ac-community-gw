# Portal characters and self-mail

## Goal

Let players see their own characters and mail items or money to them.

## Context

A player must not use the staff-wide `GET /api/v1/azeroth/characters` list. Use
cases P4 and P5 ([../../../use-cases.md](../../../use-cases.md)) require
self-scoped access with ownership enforced server-side.

## Requirements

- Backend: a self-scoped characters endpoint for the authenticated user
  (`GET /api/v1/azeroth/users/me/characters` or equivalent), resolved through the
  account link.
- Backend: a self-mail path that rejects a recipient that does not belong to the
  user's linked account.
- Portal `/characters` lists only owned characters; `/characters/$name` shows
  detail and the self-mail form.
- Without a linked account, route to `/onboarding` instead of an empty state.
- Permissions scoped to self; `azeroth.mail.send` bounded to owned characters.

## Acceptance criteria

- A player sees only their characters; requesting another user's is forbidden.
- Self-mail to a non-owned character is rejected by the API.
- New endpoints are annotated for OpenAPI and the client regenerated.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Ownership is enforced server-side; the SPA must never be the control.
- Reuse the existing character detail and mail form components from 005.

## Tests

- Go handler tests for ownership: own, other user's, and unlinked.
- RTL + MSW for the portal views.

## Dependencies

- 001. 010/011 provide onboarding for unlinked users.
