# Permissions and RBAC

Authorization is permission-based. There are no checks like `if user.IsAdmin`
scattered through the code.

```
User -> Roles -> Permissions
```

## Namespaces

Permission names are `<namespace>.<domain>...`, where the first segment is the
namespace (see [ADR 0014](ADR/0014-gateway-and-game-permission-namespaces.md)):

- **`gw`** — gateway-generic capabilities shared by every game:
  `gw.identity.user.read`, `gw.identity.roles.manage`, `gw.audit.read`,
  `gw.apikeys.manage`, `gw.notes.read`, `gw.notes.write`, `gw.notes.manage`,
  `gw.report.create`, `gw.report.read`, `gw.store.catalog.read`,
  `gw.store.wallet.read`, `gw.store.orders.read`, `gw.store.purchase`,
  `gw.store.admin.wallets`, `gw.store.admin.products`,
  `gw.store.admin.orders.read`, `gw.store.admin.orders.resolve`.
- **A registered game id** — game-specific capabilities. For AzerothCore:
  `azeroth.account.{read,manage,list,link,self}`,
  `azeroth.character.{list,self}`, `azeroth.mail.{send,self}`,
  `azeroth.leaderboard.read`, `azeroth.info.public.read`, `azeroth.item.list`,
  `azeroth.admin.claims.read`, `azeroth.admin.accounts.{ban,gmlevel}`,
  `azeroth.admin.players.{read,kick,mute}`, `azeroth.admin.characters.ban`,
  `azeroth.admin.announce`.

`permissions.Definition` declares its `Namespace`; the registry rejects a name
that does not start with `<Namespace>.`.

## Ownership

Each plugin owns its permissions. The core provides the registry, role mappings
and checks, but never knows concrete plugin permissions.

## Registering

```go
reg.Permissions.Register(permissions.Definition{
    Name:        "gw.identity.roles.manage",
    Description: "Manage roles, permission grants and Discord role mappings",
    Owner:       "identity-discord",
    Namespace:   "gw",
})
```

## Enforcing

Routes are wrapped with core middleware:

```go
reg.Mux.Handle("GET /api/v1/admin/roles",
    reg.RequirePermission("gw.identity.roles.manage", handler))
```

`RequirePermission` authenticates first (401 on failure) and then checks the
permission (403 on failure).

## Role mappings

Role -> permission grants live in the core (`permissions.Authorizer`, backed by
`roles`, `permissions` and `role_permissions`). The authorizer is loaded from
`role_permissions` at startup and **reloaded on an interval**
(`ACGW_PERMISSIONS_REFRESH_INTERVAL`, default `30s`; `0` disables it), so
changing grants in the database takes effect without a restart. The SPA also
polls `GET /api/v1/me` while the tab is focused, so the visible roles,
permissions and gated navigation update automatically.

## See also

- [ADR 0005](ADR/0005-module-owned-permissions.md)
- [ADR 0014](ADR/0014-gateway-and-game-permission-namespaces.md)
- [authentication.md](authentication.md)
