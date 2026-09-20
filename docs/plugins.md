# Plugins

A plugin is a compiled-in module that owns a slice of domain behaviour. The
contract is intentionally small.

```go
type Plugin interface {
    Name() string
    Register(ctx context.Context, reg *coreplugins.Registry) error
}
```

`Register` receives a `*plugins.Registry` with the core facilities:

```go
type Registry struct {
    Mux         *http.ServeMux
    Commands    *commands.Registry
    Services    *services.Registry
    Events      *events.Bus
    Permissions *permissions.Registry
    Audit       audit.Recorder

    RequireAuth       func(http.Handler) http.Handler
    RequirePermission func(permissions.Permission, http.Handler) http.Handler
}
```

A plugin may register, as needed:

- HTTP routes (`Mux`)
- permissions (`Permissions`)
- application commands (`Commands`)
- services/capabilities (`Services`)
- event handlers (`Events`)
- background jobs (not part of this iteration)

## Modular ownership

Each plugin owns:

- its HTTP routes
- its permissions
- its application commands
- the AzerothCore CLI syntax required by its capabilities
- the events it emits
- the services/capabilities it publishes
- its SQL queries and domain state

## Dependency rules

A plugin may import:

- `internal/core/...` contracts
- its own packages (including its generated sqlc repository)

A plugin may **not** import:

- another plugin (`internal/plugins/<other>/...`)
- an adapter (`internal/adapters/...`)

Cross-plugin interaction happens only through core registries.

## Current plugins

| Plugin | Owns (initial scope) |
| --- | --- |
| `identity-discord` | Discord identity, community users, sessions |
| `azeroth-account` | account create / password / email, account links |
| `azeroth-admin` | account ban / unban / gmlevel |
| `azeroth-info` | read-only server information |
| `azeroth-store` | store boundary, permissions and event contracts (stub) |

Planned: `azeroth-character`, `azeroth-guild`, `azeroth-support`,
`azeroth-events`, `azeroth-rewards`, `azeroth-realm`, `discord-bot`.

## Adding a plugin

1. Create `internal/plugins/<name>/plugin.go` implementing `Plugin`.
2. Add its permissions in `permissions.go` owned by the plugin name.
3. Register application commands with `commands.RegisterTyped`.
4. Publish capabilities with `services.Provide`.
5. Mount routes with `reg.Mux.Handle`, wrapping protected ones with
   `reg.RequireAuth` / `reg.RequirePermission`.
6. Wire it in `cmd/server/main.go` with `manager.Add(...)`.

No core changes are required to add a plugin.

## See also

- [architecture.md](architecture.md)
- [ADR 0008](ADR/0008-no-cross-plugin-imports.md)
