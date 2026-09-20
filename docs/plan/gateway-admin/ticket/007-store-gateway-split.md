# Split the store into a gateway plugin

## Goal

Make the store a gateway plugin so `/api/v1/store/*` and `/api/v1/admin/store/*`
are owned by a gateway plugin, as ADR 0014 requires.

## Context

`internal/plugins/azerothstore` owns the `store` permission root
(`gw.store.*`) and the gateway store routes (`/api/v1/store/*`,
`/api/v1/admin/store/*`) but is named and classified as a game plugin. The store
schema (products, wallets, orders) is gateway-level; only reward delivery is
game-specific, and it already goes through the `azeroth.character.delivery`
capability.

## Requirements

- Rename/split the plugin into a gateway `store` plugin
  (`internal/plugins/store`, `Name = "store"`) owning the `gw.store.*`
  permissions and the store routes.
- Keep reward delivery as a consumed capability; do not import the game plugin.
- Migrate `sqlc.yaml`, the generated repository package and
  `cmd/server/main.go` wiring.
- Update `docs/store.md` and any references to `azeroth-store`.

## Acceptance criteria

- The store routes are served by a gateway plugin and no longer classified as a
  game plugin by the architecture test.
- Store tests and the SPA store pages pass unchanged (URLs are unchanged).
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- This is a **roadmap** item: it is larger than the rest of the plan and does not
  block the generic-admin route move. Schedule it when a second game is added.
- The delivery capability already decouples game execution; the split is mostly
  naming, package layout and wiring.

## Tests

- Existing store handler/repository tests move with the package.

## Dependencies

- 001 (gateway plugin precedent). Independent of 002-006.
