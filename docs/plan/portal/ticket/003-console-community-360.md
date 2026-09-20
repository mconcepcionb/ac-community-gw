# Console community and 360 view

## Goal

Give staff one place to find a community user and see everything known about
them.

## Context

Today `/identity/users` and `/account-links` are separate. This ticket covers use
case C1 and prepares the ground for G6
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- `/admin/users` lists and searches community users (the existing list, rehomed).
- `/admin/users/$userId` shows the 360 view: Discord profile, roles, linked
  account, characters, wallet, orders and recent audit entries.
- The audit section is a placeholder until 013 adds the audit read capability.
- Inline actions (grant, link/unlink, ban) appear only with permission and
  delegate to their use cases rather than duplicating mutations.
- Fold `/account-links` into the user detail and the accounts area.

## Acceptance criteria

- Search works by Discord id, Discord username or community user id.
- The detail assembles existing endpoints and degrades per section when a source
  is unavailable (for example, no linked account).
- The old `/identity/users` and `/account-links` paths are removed.
- `task web:check` green.

## Implementation notes

- Query each section in parallel and render per-section loading and error states.
- Link to C2 (grant), 002 (ban) and 013 (audit) instead of reimplementing them.

## Tests

- MSW covering all sections, partial failures and permission-hidden actions.

## Dependencies

- 001. Enriched by 004 (wallet/orders summary) and 013 (audit).
