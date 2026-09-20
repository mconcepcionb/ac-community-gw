# Permissions and RBAC

Authorization is permission-based. There are no checks like `if user.IsAdmin`
scattered through the code.

```
User -> Roles -> Permissions
```

## Ownership

Each plugin owns its permissions. The core provides the registry, role mappings
and checks, but never knows concrete plugin permissions.

Examples:

```
identity.self.read
identity.session.revoke

azeroth.account.read
azeroth.account.manage

azeroth.admin.accounts.read
azeroth.admin.accounts.ban
azeroth.admin.accounts.gmlevel

azeroth.info.public.read
azeroth.info.private.read

azeroth.store.read
azeroth.store.manage
```

## Registering

```go
reg.Permissions.Register(permissions.Definition{
    Name:        "azeroth.account.manage",
    Description: "Create and manage AzerothCore accounts",
    Owner:       "azeroth-account",
})
```

## Enforcing

Routes are wrapped with core middleware:

```go
reg.Mux.Handle("GET /api/v1/foo",
    reg.RequirePermission("azeroth.account.read", handler))
```

`RequirePermission` authenticates first (401 on failure) and then checks the
permission (403 on failure).

## Role mappings

Role -> permission grants live in the core (`permissions.Authorizer`, backed by
`roles`, `permissions` and `role_permissions`). The first iteration uses an
in-memory authorizer; a PostgreSQL-backed implementation can replace it without
changing the middleware contract.

The authorizer is loaded from `role_permissions` at startup and **reloaded on an
interval** (`ACGW_PERMISSIONS_REFRESH_INTERVAL`, default `30s`; `0` disables it),
so changing grants in the database takes effect without a restart. The SPA also
polls `GET /api/v1/me` while the tab is focused, so the visible roles,
permissions and gated navigation update automatically.

## See also

- [ADR 0005](ADR/0005-module-owned-permissions.md)
- [authentication.md](authentication.md)
