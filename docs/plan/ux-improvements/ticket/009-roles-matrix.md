# Roles: mappings tab and permission matrix

## Goal

Replace the free-text roles page with two focused tabs: Discord role mappings
and a role/permission matrix with checkboxes and an explicit save.

## Context

`web/src/features/admin/admin-roles-page.tsx` is a single page with free-text
`Role`/`Permission` inputs and an immediate grant/revoke per interaction
(`:142-164`, `:194-216`).

The backend cannot currently power a matrix:

- `GET /api/v1/admin/roles` returns only existing `grants` and `mappings`
  (`internal/plugins/identitydiscord/rolesadmin.go:29-32, 56-85`).
- There is no roles list endpoint; `Authorizer.RoleNames()` exists but is unused
  (`internal/core/permissions/rbac.go:82`).
- The permission catalog is at `GET /api/v1/admin/permissions`, gated by
  `gw.apikeys.manage` (`internal/plugins/apikeys/plugin.go:59-60`); 004 opens it
  to `gw.identity.roles.manage`.

## Requirements

- Backend: extend (or add) `GET /api/v1/admin/roles` to return
  `{ roles: string[], permissions: [{name, description, owner}], grants, mappings }`,
  or consume a catalog endpoint made available by 004.
- Backend: replace the per-item grant/revoke endpoints with a batch save
  endpoint (e.g. `PUT /api/v1/admin/roles/{role}/permissions` with the full
  desired set) that diffs against current grants in one transaction. The
  project is pre-production, so the old endpoints are removed rather than kept
  for compatibility.
- Create/delete role endpoints (or explicit "create role" in the save).
- Console `/admin/roles` two tabs (use `components/ui/tabs.tsx`):
  1. **Discord mappings** — table of `discord_role_id → role` with add/edit
     dialog and delete confirmation.
  2. **Role/permission matrix** — permissions grouped by namespace (Gateway,
     then each game) and owner, one column (or section) per role, checkbox per
     permission, sticky **Save** that submits the diff; confirmation on
     revokes.
- Matrix shows description/owner metadata and a "0 permissions" role.

## Acceptance criteria

- The matrix renders every registered permission with checkboxes reflecting
  current grants.
- Saving applies grants and revokes atomically and refreshes effective access
  within the refresh interval.
- Destructive changes are confirmed and audited; self-escalation remains
  impossible.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- Reuse the permission checkbox pattern from
  `admin-api-clients-page.tsx:114-131`.
- Reuse the authorizer reload; no new propagation.
- Keep role names free-form at the API but offer existing roles in the UI to
  avoid typos.

## Tests

- Go tests for catalog return, batch diff, authorization and audit.
- RTL + MSW for both tabs, including the save diff.

## Dependencies

- 001, 004 (catalog access), 017 (existing roles admin, historical).
