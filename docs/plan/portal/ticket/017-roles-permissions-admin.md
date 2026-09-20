# Roles and Discord-role mapping administration

## Goal

Let staff manage internal roles, permission grants and Discord-role mappings from
the console.

## Context

Today this is environment configuration plus SQL; the authorizer already reloads
from the database on an interval. This ticket covers use case C4
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- Backend: CRUD for role-to-permission grants and Discord-role mappings.
- Changes reach signed-in users within the existing refresh interval, without a
  restart.
- Console `/admin/roles`: manage mappings and grants, with confirmation on
  destructive changes.
- All changes audited; a new administration permission.

## Acceptance criteria

- A new mapping changes effective access within the refresh interval.
- Destructive changes require confirmation and are audited.
- Only permitted staff can manage roles; self-escalation is impossible.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Reuse the existing authorizer reload mechanism; no new propagation needed.
- Validate that a role cannot grant itself permissions it does not hold.

## Tests

- Go tests for authorization and audit; RTL + MSW for the admin UI.

## Dependencies

- 001, 002. Relates to 016 (auditor role) and 020 (API-key scopes).
