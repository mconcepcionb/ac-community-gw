# Architecture

`ac-community-gw` is a **modular monolith**. Every plugin is compiled into a
single binary; there is no dynamic loading, no RPC and no service mesh.

```
Consumers (web, bots)
        |
        | REST / HTTPS
        v
  ac-community-gw  (core + plugins)
        |
        | controlled internal integrations
        v
AzerothCore SOAP   Discord   PostgreSQL
```

## Layers

```
cmd/server            wiring, lifecycle
internal/core         transversal infrastructure, no domain knowledge
internal/plugins      domain modules (compiled in)
internal/adapters     external integrations (PostgreSQL, AzerothCore SOAP)
migrations            single schema lifecycle
```

`cmd/server` is the only place that imports both `core` and the concrete
plugins; it wires everything together.

## Dependency rules

```
core MUST NOT import plugins
plugins MAY import core contracts
a plugin MUST NOT import another plugin's implementation
core and plugins MUST NOT import adapters
```

These rules are mechanically enforced by
`internal/architecture/architecture_test.go`, which fails the build if a
forbidden import appears.

## Core responsibilities

The core provides infrastructure only:

- configuration
- HTTP server and middleware (request id, recovery, access log, auth, RBAC)
- application lifecycle
- plugin registry
- authentication primitives and sessions
- authorization / RBAC and permission registry
- command registry
- service / capability registry
- in-process synchronous event bus
- PostgreSQL infrastructure and migrations
- audit infrastructure
- health and readiness
- structured logging (`log/slog`)

The core does **not** contain AzerothCore domain logic and has no hardcoded
catalogue of permissions, commands or events.

## Cross-plugin communication

Four mechanisms, deliberately distinct:

| Need | Mechanism | Registry |
| --- | --- | --- |
| Synchronous read/validation | **Query** | service / capability registry |
| Request an action | **Command** | application command registry |
| Announce a fact that happened | **Event** | in-process event bus |
| Check authorization | **Permission** | permission registry + RBAC |

- Events are never used as commands or queries.
- Commands are never used as a generic query mechanism.

Because each plugin publishes and consumes only by name and Go interface,
plugins stay decoupled and no import edge forms between them.

## Metadata hygiene

- Events carry identifiers, never secrets or sensitive payloads.
- Audit entries never contain passwords, tokens, session IDs or credentials.
- Plugins emit domain events defined in their own package; consumers declare
  the interface or event name they care about locally.
- Only events with a real publisher and consumer exist; unreferenced event types
  are removed rather than kept as dead code (the identity-discord events are the
  only ones currently published).

## Lifecycle

1. Load configuration from the environment.
2. Build the structured logger.
3. Open PostgreSQL (if configured) and optionally run Goose migrations.
4. Build core registries (commands, services, events, permissions, audit).
5. Build the HTTP server.
6. Register every plugin against the core registry.
7. Serve traffic until `SIGINT`/`SIGTERM`, then shut down gracefully.

## See also

- [ADR 0001](ADR/0001-modular-monolith.md)
- [ADR 0008](ADR/0008-no-cross-plugin-imports.md)
- [ADR 0009](ADR/0009-core-command-service-event-registries.md)
