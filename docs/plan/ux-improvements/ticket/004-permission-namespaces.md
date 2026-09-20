# Gateway/game permission namespaces

## Goal

Introduce the `gw.` / game-id namespace split ([ADR 0014](../../../ADR/0014-gateway-and-game-permission-namespaces.md)),
hard-rename existing permissions, move generic permissions out of the game
plugin, expose the catalog to role administrators, and remove dead permissions.

## Context

The gateway will grow to more games, so the boundary between gateway-generic and
game-specific permissions and routes must be explicit (ADR 0014). Current drift:

- User list is gated by `identity.user.list`
  (`internal/plugins/identitydiscord/permissions.go:11`); user detail/360 by
  `azeroth.admin.users.read` (`internal/plugins/azerothadmin/permissions.go:22`)
  — a generic read under a game root and a different permission.
- `audit.read` is registered and owned by `azeroth-admin`
  (`internal/plugins/azerothadmin/permissions.go:24`), a generic concern.
- Generic routes are served by the game plugin: `GET /api/v1/admin/users/{id}`
  and `GET /api/v1/admin/audit` in `azerothadmin/plugin.go:148-151`.
- The demo admin role excludes every `azeroth.admin.*` permission
  (`scripts/seed_demo_admin.sql:77-81`), so it can list users but gets 403 on
  the detail — the reported "admins should be able to see users" bug.
- The permission catalog is only served by `GET /api/v1/admin/permissions`,
  owned by `apikeys` and gated by `apikeys.manage`
  (`internal/plugins/apikeys/plugin.go:59-60`), so the roles page cannot build a
  matrix without that permission.
- Registered but never enforced: `azeroth.admin.accounts.read`,
  `azeroth.info.private.read`, `identity.self.read`, `identity.session.revoke`.
- Docs drift: `docs/permissions.md:31-32` lists non-existent `azeroth.store.*`;
  plan docs reference `identity.users.list`.

## Requirements

- Add `Namespace` to `permissions.Definition`
  (`internal/core/permissions/registry.go:29`) and validate at registration that
  `Name` starts with `<Namespace>.`; a mismatch fails registration (and CI).
- **Hard-rename** every permission per the mapping below. No aliases or
  compatibility shims — pre-production.
- Move ownership of generic permissions out of `azeroth-admin`: `gw.audit.read`
  and the community-user read belong to gateway plugins.
- Introduce `gw.identity.user.read`; enforce it on both
  `GET /api/v1/identity/users` and `GET /api/v1/admin/users/{id}`.
- Provide the catalog to role administrators: relax
  `GET /api/v1/admin/permissions` to also accept `gw.identity.roles.manage`, or
  add a gateway-owned catalog endpoint. Decide and document.
- Resolve dead permissions: enforce or remove.
- Update `docs/permissions.md`, plan docs, `web/src/features/auth/surfaces.ts`,
  `console-nav.tsx`, `scripts/seed_demo_admin.sql`, and the SPA tests.
- Record the route-ownership target (generic `/admin/*`, `/identity/*` served by
  gateway plugins) and leave the physical move as a documented follow-up.

## Mapping

### Generic → `gw.*`

| Current | Target |
| --- | --- |
| `identity.self.read` | `gw.identity.self.read` (enforce or remove) |
| `identity.session.revoke` | `gw.identity.session.revoke` (enforce or remove) |
| `identity.user.list` | `gw.identity.user.read` |
| `azeroth.admin.users.read` | `gw.identity.user.read` (removed from `azeroth-admin`) |
| `identity.roles.manage` | `gw.identity.roles.manage` |
| `report.create` | `gw.report.create` |
| `report.read` | `gw.report.read` |
| `store.catalog.read` | `gw.store.catalog.read` |
| `store.wallet.read` | `gw.store.wallet.read` |
| `store.orders.read` | `gw.store.orders.read` |
| `store.purchase` | `gw.store.purchase` |
| `store.admin.wallets` | `gw.store.admin.wallets` |
| `store.admin.products` | `gw.store.admin.products` |
| `store.admin.orders.read` | `gw.store.admin.orders.read` |
| `store.admin.orders.resolve` | `gw.store.admin.orders.resolve` |
| `audit.read` | `gw.audit.read` (moves to gateway ownership) |
| `apikeys.manage` | `gw.apikeys.manage` |
| _(new, ticket 002)_ | `gw.notes.manage` |

### Game → `azeroth.*` (root unchanged)

`azeroth.account.{read,manage,list,link,self}`, `azeroth.admin.claims.read`,
`azeroth.character.{list,self}`, `azeroth.mail.{send,self}`,
`azeroth.leaderboard.read`, `azeroth.info.public.read`, `azeroth.item.list`,
`azeroth.admin.accounts.{ban,gmlevel}`, `azeroth.admin.players.{read,kick,mute}`,
`azeroth.admin.characters.ban`, `azeroth.admin.announce`, and the new
`azeroth.admin.mail.send` (ticket 008).

### Removed (registered but never enforced)

`azeroth.admin.accounts.read`, `azeroth.info.private.read` — delete the constant
and definition unless a use is added in the same change.

## Acceptance criteria

- Every registered permission starts with `gw.` or a registered game id; the
  registry rejects a mismatch.
- A role holding `gw.identity.user.read` can list and open community users.
- The roles page can load the permission catalog with only
  `gw.identity.roles.manage`.
- No registered permission is unused; no documented permission is missing.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- Delete old permission definitions in the same change and update every
  reference. No deprecation window.
- Keep permission ownership with the plugin that enforces it; generic ones move
  to gateway plugins.
- Rename in one mechanical pass; the registry prefix check makes any missed
  reference fail fast.

## Tests

- Registry unit test: a definition whose name does not match its `Namespace`
  fails registration.
- Go tests asserting `GET /admin/users/{id}` accepts `gw.identity.user.read` and
  rejects unprivileged callers.
- Update `use-permissions`/`surfaces` tests and the user-detail page test.

## Dependencies

- Independent; implements ADR 0014. Blocks the consistency goals of 005 and the
  catalog need of 009.
