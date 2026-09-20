# ADR 0014: Gateway and game permission namespaces

## Status

Accepted

## Context

The gateway is expected to grow beyond AzerothCore and serve more than one game.
Today permissions are namespaced by domain only ([ADR 0005](0005-module-owned-permissions.md)):
some are gateway-generic (`identity.*`, `store.*`, `report.*`) while others are
AzerothCore-specific (`azeroth.*`), but nothing distinguishes the two families.
The result is drift:

- Generic permissions live under a game root (`azeroth.admin.users.read`).
- Generic permissions are owned by a game plugin (`audit.read` and
  `azeroth.admin.users.read` are registered by `azeroth-admin`).
- Depth is inconsistent (`identity.user.list` vs `audit.read`).
- A generic route namespace (`/api/v1/admin/*`) is served by a game plugin
  (`GET /api/v1/admin/users/{id}`, `GET /api/v1/admin/audit`).

As more games are added, the boundary between "all of the gateway" and "this
game" must be explicit and mechanically enforced, not left to convention.

## Decision

Permission names are `<namespace>.<domain>...`, where the **first segment is the
namespace** and is one of:

- **`gw`** — gateway-generic, game-agnostic capabilities (`gw.identity.*`,
  `gw.store.*`, `gw.report.*`, `gw.audit.*`, `gw.apikeys.*`, `gw.notes.*`).
- **A registered game id** — game-specific capabilities (`azeroth.account.*`,
  `azeroth.character.*`, `azeroth.item.*`, `azeroth.mail.*`, `azeroth.info.*`,
  `azeroth.leaderboard.*`, `azeroth.admin.*`).

The namespace is declared, not inferred: `permissions.Definition` gains a
`Namespace` field and the permission registry validates at registration time
that the name starts with `<namespace>.`. A mismatch fails the build.

The same split applies to HTTP routes:

- Gateway routes: `/api/v1/identity/*`, `/api/v1/store/*`, `/api/v1/admin/*`,
  `/api/v1/public/*`, `/api/v1/me`, `/api/v1/auth/*`.
- Game routes: `/api/v1/<game>/*` (`/api/v1/azeroth/*`).

A gateway route or permission is owned by a gateway plugin; a game route or
permission is owned by that game's plugin. Generic aggregates that compose game
capabilities (for example a multi-game user 360) live in a gateway plugin and
read game data only through the core service/capability registry.

`gw` is reserved; no game may be named `gw`. Game ids are the plugin namespace
they belong to.

## Consequences

- The boundary is visible in every permission string and route.
- The registry enforces the split, so drift fails CI rather than review.
- Existing generic permissions are hard-renamed to `gw.*`; the project is
  pre-production, so no compatibility aliases are kept.
- The roles/permission matrix can group the catalog by namespace (Gateway vs
  each game).
- Generic permissions currently owned by `azeroth-admin` (`audit.read`,
  `azeroth.admin.users.read`) move to gateway ownership.
- Serving generic routes from game plugins is a known follow-up; the ADR states
  the target so new work does not repeat the mistake.
- The core still holds no catalogue of concrete permissions
  ([ADR 0005](0005-module-owned-permissions.md)); it only holds the `Namespace`
  field and the prefix check.

## See also

- [ADR 0005](0005-module-owned-permissions.md)
- [ADR 0006](0006-module-owned-azeroth-commands.md)
- [../plan/ux-improvements/ticket/004-permission-namespaces.md](../plan/ux-improvements/ticket/004-permission-namespaces.md)
