# Console community and 360 view

## Goal

Give staff one page that shows everything known about a community user, served
by a dedicated admin endpoint rather than player-scoped reads.

## Context

Today `/identity/users` and `/account-links` are separate, and
`store.orders.read` / `store.wallet.read` are explicitly **self-scoped**
(`internal/plugins/azerothstore/permissions.go`), so they cannot back a staff
360 view. This ticket covers use case C1 and prepares G6
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- Backend: a single aggregate endpoint
  `GET /api/v1/admin/users/{id}` under a dedicated staff permission, returning
  the Discord profile, roles, linked account, characters, wallet, orders and an
  audit summary.
- `/admin/users` lists and searches community users.
- `/admin/users/$userId` renders the 360 view; the audit section is a placeholder
  until 016.
- Inline actions (grant, link/unlink, ban) appear only with permission and
  delegate to their use cases rather than duplicating mutations.
- Fold `/account-links` into the user detail and the accounts area.

## Acceptance criteria

- The aggregate returns every section in one response; sections that are
  temporarily unavailable degrade individually, not as a whole.
- Search works by Discord id, Discord username or community user id.
- The view never uses self-scoped player endpoints.
- Endpoints are annotated and the client regenerated.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- The aggregate is composed in the backend from the owning capabilities, so the
  SPA makes one call.
- Reuse `azeroth.account.directory` and the store reads server-side.

## Tests

- Go aggregate handler tests (permission, partial source failure).
- RTL + MSW for the list, detail and permission-hidden actions.

## Dependencies

- 001, 002. Enriched by 005 (store summary) and 016 (audit).
