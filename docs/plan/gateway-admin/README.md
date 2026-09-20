# Gateway admin route ownership

## Goal

Complete the route-ownership target of
[ADR 0014](../../ADR/0014-gateway-and-game-permission-namespaces.md): generic
gateway routes (`/api/v1/admin/*`, `/api/v1/identity/*`, `/api/v1/store/*`,
`/api/v1/public/*`) are served by gateway plugins, and game-specific routes live
under `/api/v1/<game>/*`. Today several generic routes are still served by game
plugins (and vice versa), which will not survive a second game.

## Status

| Area | State |
| --- | --- |
| Route inventory and classification | done |
| Gateway admin plugin (skeleton, wiring, ownership) | implemented (001) |
| Move community user 360 into the gateway plugin | implemented (002) |
| Move the audit viewer and `gw.audit.read` | implemented (001, 003) |
| Correct game routes wrongly under `/admin` | implemented (004) |
| Decide the `/public/*` surface ownership | implemented (005); game public data moved under `/azeroth/public/*`, gateway aggregator deferred |
| Move the permission catalog out of `apikeys` | implemented (006) |
| SPA console route and nav split (core vs per-game) | planned |
| SPA portal route split (core vs per-game) | planned |
| Split `azeroth-store` into a gateway `store` plugin | roadmap |

## Context

ADR 0014 fixes the namespace split for permissions and routes. Permission
namespaces were implemented (see [../ux-improvements/ticket/004-permission-namespaces.md](../ux-improvements/ticket/004-permission-namespaces.md));
route and plugin ownership was left as a follow-up. The full route inventory
shows the remaining drift:

| Route | Served by | Kind | Problem |
| --- | --- | --- | --- |
| `GET /api/v1/admin/users/{id}` | `azeroth-admin` | game plugin | generic route in a game plugin |
| `GET /api/v1/admin/audit` | `azeroth-admin` | game plugin | generic route in a game plugin |
| `GET /api/v1/admin/account-claims` | `azeroth-account` | game plugin | generic route namespace for game data |
| `POST /api/v1/admin/characters/{name}/mail` | `azeroth-character` | game plugin | generic route namespace for game data |
| `GET /api/v1/admin/permissions` | `apikeys` | gateway plugin | catalog is not API-key specific |
| `GET /api/v1/admin/store/orders` (+ refund/retry) | `azeroth-store` | game-named plugin | `store` is a gateway root |
| `GET /api/v1/public/leaderboards/{board}` | `azeroth-character` | game plugin | generic `/public` surface for game data |
| `GET /api/v1/public/status` | `azeroth-info` | game plugin | generic `/public` surface for game data |

Correctly owned today: `/api/v1/admin/roles` and `/api/v1/admin/discord-role-mappings`
(`identity-discord`), `/api/v1/admin/annotations` (`admin-notes`),
`/api/v1/admin/reports` (`reports`), `/api/v1/admin/api-keys` (`apikeys`),
`/api/v1/identity/*` (`identity-discord`).

The user 360 aggregate is the hard case: it composes identity, account,
character and store data. It already reads game data through the core
service/capability registry, so moving it to a gateway plugin is a wiring change,
not a data change.

## Scope

- A new gateway plugin that owns the generic admin reads (`/admin/users/{id}`,
  `/admin/audit`, and the permission catalog).
- Moving the user-360 aggregate and the audit viewer into it, with
  `gw.audit.read` ownership.
- Correcting game-specific routes that currently sit under `/admin` so they live
  under `/api/v1/azeroth/admin/*`.
- Deciding and implementing the `/public/*` surface ownership.
- Renaming the generated OpenAPI operation ids and the SPA call sites where the
  route moves.
- Splitting the SPA route tree and navigation into gateway-core and per-game
  sections, mirroring the API.
- Keeping `task check` / `task web:check` green at every step.

## Out of scope

- Changing the RBAC model, the session model or the serving model.
- A multi-game data model for the 360 (per-game nesting). The move preserves the
  current response shape; a multi-game shape is a separate decision.
- The store's product/wallet/order schema. The gateway split is a follow-up.

## Design

### Ownership rules

- **A gateway route is served by a gateway plugin.** Gateway routes are
  `/api/v1/admin/*`, `/api/v1/identity/*`, `/api/v1/store/*`,
  `/api/v1/public/*`, `/api/v1/me`, `/api/v1/auth/*`.
- **A game route is served by a game plugin and lives under
  `/api/v1/<game>/*`.** Game-specific admin actions are
  `/api/v1/<game>/admin/*`.
- **A permission's namespace matches its owner's kind**, not the route it gates.
  A game permission may gate a game route; a gateway route uses a gateway
  permission.

### SPA route conventions

The SPA mirrors the API split. The SPA is not deployed, so routes are moved
outright (no redirects).

- **Console core** stays under `/admin/*`: `/admin` (overview), `/admin/users`,
  `/admin/roles`, `/admin/audit`, `/admin/api-clients`, `/admin/moderation`,
  `/admin/store`.
- **Console per-game** moves under `/admin/<game>/*`:
  `/admin/azeroth/accounts`, `/admin/azeroth/characters`,
  `/admin/azeroth/items`, `/admin/azeroth/online`.
- **Portal core** stays at the root: `/`, `/login`, `/profile`, `/wallet`,
  `/store`, `/report`, `/forbidden`.
- **Portal per-game** moves under `/<game>/*`: `/azeroth/characters`,
  `/azeroth/leaderboards`, `/azeroth/status`, `/azeroth/onboarding`.
- The game id is a single constant (`azeroth` today) so a second game adds a
  route subtree rather than rewriting links.
- The console nav and the overview cards group core links first, then a
  game-labelled section.

### The gateway admin plugin

A new plugin `gateway-admin` (`internal/plugins/gatewayadmin`) owns the generic
console aggregates:

- `GET /api/v1/admin/users/{id}` — community user 360.
- `GET /api/v1/admin/audit` — audit viewer.
- `GET /api/v1/admin/permissions` — registered permission catalog (moved from
  `apikeys`).

It is a gateway plugin, so it:

- owns `gw.audit.read` (moved from `azeroth-admin`);
- consumes game capabilities (`azeroth.account.directory`,
  `azeroth.character.directory`, `azeroth.store.account`) and
  `identity.user.admin` **lazily** through the core service registry, exactly as
  `azeroth-admin` does today.

The URL of the moved routes does not change; only the serving plugin does. The
OpenAPI operation ids change from `azeroth.admin.*` to `gateway.admin.*`, which
renames the generated TS client functions.

### Multi-game direction (noted, not delivered)

The 360 response keeps its current flat shape in this plan. The multi-game shape
(a `games` object keyed by game id, each with its own account/characters) is a
future decision and is not required for the ownership move.

### `/public/*`

`/public/*` is a gateway surface. Game-specific public data must not be served
there by a game plugin. Two options:

1. **Game prefix** (small): expose game public data under
   `/api/v1/azeroth/public/*` (leaderboards, status). A future gateway `public`
   plugin can aggregate.
2. **Gateway public plugin** (larger): a gateway `public` plugin owns
   `/api/v1/public/*` and composes per-game capabilities.

This plan recommends **option 1** now and leaves the aggregating gateway public
plugin as a follow-up; it keeps the change mechanical and preserves the
no-login surface.

### Backend gaps

| Gap | Needed by |
| --- | --- |
| `gateway-admin` plugin skeleton and wiring | 001 |
| `gw.audit.read` ownership move | 001 |
| Lazy capability resolution for the 360 | 002 |
| Audit reader injected into the gateway plugin | 003 |
| Permission catalog endpoint moved | 006 |

### Milestones

| Milestone | Tickets |
| --- | --- |
| A - Gateway admin plugin | 001-003 |
| B - Route namespace corrections | 004-005 |
| C - Catalog and store | 006-007 |
| D - SPA route split | 008-009 |

### Dependency graph

```
001 -> 002, 003, 006
002, 003, 004, 005 -> 008, 009
008, 009 -> (nav and overview updates)
```

001 creates the plugin and moves `gw.audit.read`; 002 and 003 move the two
routes into it. 004/005 fix the game routes. 006 moves the permission catalog.
008/009 restructure the SPA after the API paths settle, so the call-site updates
and the route moves land once.

## Risks

| Risk | Mitigation |
| --- | --- |
| The 360 loses game data when moved | It already resolves game capabilities lazily; move the resolution code verbatim and cover it with the existing user-360 tests |
| Operation-id rename breaks the SPA | Rename `@ID`s and regenerate the client in the same ticket; update the SPA call sites (`azerothAdminUsersGet`, `azerothAdminAuditList`, `azerothAdminAccountClaimsList`, `azerothAdminCharactersMail`) |
| Route moves break deep links | The SPA is not deployed, so routes move outright; the API `/admin/users` and `/admin/audit` keep their URLs |
| Two plugins briefly own `gw.audit.read` | Remove the old definition in the same change; the registry rejects duplicates |
| Gateway plugin becomes a god object | Keep it to console aggregates; no game domain logic, only capability composition |

## Tickets

### Milestone A - Gateway admin plugin

1. [001-gateway-admin-plugin.md](ticket/001-gateway-admin-plugin.md) - plugin skeleton, wiring and `gw.audit.read` ownership
2. [002-move-user-360.md](ticket/002-move-user-360.md) - move the community user 360
3. [003-move-audit-viewer.md](ticket/003-move-audit-viewer.md) - move the audit viewer

### Milestone B - Route namespace corrections

4. [004-game-route-namespaces.md](ticket/004-game-route-namespaces.md) - move account-claims and character mail under the game prefix
5. [005-public-surface.md](ticket/005-public-surface.md) - move game public data under the game prefix

### Milestone C - Catalog and store

6. [006-permission-catalog.md](ticket/006-permission-catalog.md) - move the permission catalog out of `apikeys`
7. [007-store-gateway-split.md](ticket/007-store-gateway-split.md) - split the store into a gateway plugin (roadmap)

### Milestone D - SPA route split

8. [008-spa-console-split.md](ticket/008-spa-console-split.md) - core vs per-game console routes and nav
9. [009-spa-portal-split.md](ticket/009-spa-portal-split.md) - core vs per-game portal routes
