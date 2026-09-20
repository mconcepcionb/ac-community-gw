# Permission catalog ownership

## Goal

Give the registered permission catalog a semantically correct gateway owner.

## Context

`GET /api/v1/admin/permissions` is served by `apikeys`
(`internal/plugins/apikeys/plugin.go:59`) and gated by `gw.apikeys.manage`. The
route is a gateway route served by a gateway plugin, so ADR 0014 is already
satisfied; the only concern is that the catalog is not API-key specific.

The catalog has a single SPA consumer today: the API-clients scope picker
(`admin-api-clients-page.tsx`). The roles matrix no longer needs it — it reads
the catalog embedded in `GET /api/v1/admin/roles` (added in
[../../ux-improvements/ticket/009-roles-matrix.md](../../ux-improvements/ticket/009-roles-matrix.md)).

## Requirements

- Move the endpoint to `gateway-admin` as
  `GET /api/v1/admin/permissions` with `@ID gateway.admin.permissions.list`.
- Gate it with `gw.apikeys.manage` (unchanged) so the API-clients page keeps
  working, or introduce a gateway permission if a broader consumer appears.
- Regenerate the client; update the `apikeysPermissionsList` call site in
  `admin-api-clients-page.tsx`.

## Acceptance criteria

- The catalog is served by the gateway admin plugin.
- The API-clients page loads the catalog.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- Low priority: this is a semantic cleanup, not a rule violation. It can be
  dropped if the gateway admin plugin grows too large.
- Do not add OR-permission support to the middleware for a single endpoint.

## Tests

- Update the handler test and the API-clients page test that mocks the catalog
  URL.

## Dependencies

- 001. Optional; can land any time after 001.
