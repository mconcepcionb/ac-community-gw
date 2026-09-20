# Move the community user 360

## Goal

Serve `GET /api/v1/admin/users/{id}` from the gateway admin plugin instead of
`azeroth-admin`, preserving the aggregate and its URL.

## Context

The 360 handler (`internal/plugins/azerothadmin/user360.go`) composes
`identity.user.admin`, `azeroth.account.directory`,
`azeroth.character.directory` and `azeroth.store.account` through the service
registry. The route is generic but is served by a game plugin.

## Requirements

- Move `user360.go` and its tests into `gateway-admin`.
- Resolve the four capabilities lazily via the stored service registry (as
  `azeroth-admin` does today); do not import sibling plugins.
- Keep the URL `GET /api/v1/admin/users/{id}`.
- Rename the operation id `azeroth.admin.users.get` to `gateway.admin.users.get`.
- Enforce `gw.identity.user.read` (unchanged).
- Remove the now-unused `permissionUserRead` constant from `azeroth-admin`.
- Regenerate the OpenAPI spec and TS client; update the SPA
  (`azerothAdminUsersGet` → `gatewayAdminUsersGet`) and its tests.

## Acceptance criteria

- The user detail page and its test pass against the new owner.
- No `azeroth.admin.users.*` operation id remains in `api/swagger.yaml`.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- Move the `AdminUser`/`AdminUserCharacter`/`AdminUserOrder` DTOs with the
  handler; keep `@name` values so schema names do not churn unnecessarily.
- The 360 keeps its flat response shape in this ticket; multi-game nesting is out
  of scope (see the plan README).

## Tests

- Relocate `user360_test.go`; it already exercises the capability composition
  with fakes.

## Dependencies

- 001. Consumed by the SPA user detail page.
