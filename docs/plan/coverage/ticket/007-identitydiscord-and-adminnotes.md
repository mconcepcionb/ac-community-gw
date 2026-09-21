# identitydiscord and adminnotes

**Milestone:** B · **Gate:** G-standard · **Depends on:** 001, 003

## Goal

Raise `identitydiscord` (50.7%) and `adminnotes` (35.7%) to ≥ 60%.

## Context

- `identitydiscord`: `rolesadmin.go` (roles list, permission catalog, batch
  grant save, Discord-role mappings), `provision.go`, `response.go`
  (`effectivePermissions`), `janitor.go` (`RunCleanup`), `events.go`,
  `permissions.go` and `Register` are untested
  (`internal/plugins/identitydiscord/`). OAuth, sessions, users and roles sync
  have tests.
- `adminnotes`: `handleDelete`, the create happy path, `canManage`, audit
  `record`, `permissionDefs` and `Register` are untested
  (`internal/plugins/adminnotes/handlers.go`).

## Requirements

- `identitydiscord`: role/permission admin happy and error paths; the permission
  catalog; `effectivePermissions` merging; provisioning; janitor cleanup.
- `adminnotes`: delete (including authorization), create success, `canManage`
  matrix, audit record shape.
- Cover both `Register`s and permission definitions.

## Tests

- A role-permission batch save test with an invalid permission rejected.
- An adminnotes delete-by-non-author/-non-moderator test (`403`).

## Acceptance criteria

- Both plugins reach ≥ 60%.
- `task check` green.

## Out of scope

- The identitydiscord repository (C9, integration).
