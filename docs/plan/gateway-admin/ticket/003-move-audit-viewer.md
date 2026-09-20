# Move the audit viewer

## Goal

Serve `GET /api/v1/admin/audit` from the gateway admin plugin.

## Context

`internal/plugins/azerothadmin/auditview.go` exposes the audit viewer and is
gated by `gw.audit.read`, which ticket 001 moves to `gateway-admin`.

## Requirements

- Move `auditview.go` and `audit_test.go` into `gateway-admin`.
- Use the `audit.Reader` injected in ticket 001.
- Keep the URL `GET /api/v1/admin/audit` and the target/metadata filters added in
  [../../ux-improvements/ticket/003-audit-targeting.md](../../ux-improvements/ticket/003-audit-targeting.md).
- Rename the operation id `azeroth.admin.audit.list` to
  `gateway.admin.audit.list`.
- Remove the route and the `auditReader` field from `azeroth-admin`.
- Regenerate the client; update the SPA (`azerothAdminAuditList` →
  `gatewayAdminAuditList`) in `admin-audit-page.tsx` and `entity-history.tsx`.

## Acceptance criteria

- The audit page and the per-entity history panel pass against the new owner.
- `azeroth-admin` no longer references the audit reader.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- `audit.Reader` is a core interface, so the move has no cross-plugin coupling.
- Keep the `AdminAuditEntry`/`AdminAuditResponse` `@name`s to limit schema churn.

## Tests

- Relocate `audit_test.go`; it covers the filter/metadata wiring.

## Dependencies

- 001. Consumed by the SPA audit page and entity history.
