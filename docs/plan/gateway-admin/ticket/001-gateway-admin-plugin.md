# Gateway admin plugin

## Goal

Create the gateway plugin that will own the generic console aggregates, and move
the `gw.audit.read` permission definition out of `azeroth-admin`.

## Context

`azeroth-admin` is a game plugin (`internal/plugins/azerothadmin`) but it
registers two gateway routes and the `gw.audit.read` permission
(`internal/plugins/azerothadmin/permissions.go:24`). ADR 0014 requires generic
routes and permissions to be owned by gateway plugins.

## Requirements

- New plugin `gateway-admin` under `internal/plugins/gatewayadmin` with
  `Name = "gateway-admin"`.
- Config carries `Audit audit.Recorder` and `AuditReader audit.Reader`.
- Register `gw.audit.read` here (Namespace `"gw"`) and remove it from
  `azeroth-admin` (`PermissionAdminAuditRead`).
- Wire the plugin in `cmd/server/main.go` with `manager.Add`.
- No game domain logic; only console aggregates and capability composition.

## Acceptance criteria

- The plugin registers with no duplicate-permission error.
- `azeroth-admin` no longer defines `gw.audit.read`.
- The architecture test (`internal/architecture`) classifies and accepts the new
  plugin package.
- `task check` green.

## Implementation notes

- Start the plugin with `Register` registering only permissions and storing the
  audit reader; routes arrive in tickets 002 and 003.
- Store `reg.Services` on the plugin so later tickets can resolve game
  capabilities lazily, mirroring `azerothadmin.Plugin.registry`.

## Tests

- Registry unit coverage is already provided by the core; add a small
  `plugin_test.go` asserting `permissionDefs()` includes `gw.audit.read` with
  `Namespace: "gw"`.

## Dependencies

- Implements [ADR 0014](../../../ADR/0014-gateway-and-game-permission-namespaces.md).
  Blocks 002, 003 and 006.
