# Foundation

## Goal

Establish the operational base of `ac-community-gw`: project structure,
architecture, common infrastructure, toolchain, persistence, documentation,
tests and a minimal vertical slice.

## Context

The gateway needs a coherent start before any business capability is built. The
base must compile, start, connect to PostgreSQL, run migrations, generate typed
SQL, load plugins, serve HTTP and expose health/readiness.

## Scope

- project structure and module wiring
- core infrastructure: config, logging, HTTP, registries, event bus, RBAC,
  sessions primitives, audit, persistence contracts
- SOAP adapter and `CommandExecutor` contract
- initial plugin packages
- migrations, sqlc configuration and generated code
- Taskfile, Dockerfile, Compose, environment template
- documentation structure (stable docs, ADRs, plans, runbooks)
- unit tests and architectural boundary tests

## Out of scope

- complete store functionality
- complete account administration
- all account endpoints
- full role synchronisation
- Discord bot, guilds, rewards, tickets, economy

## Design

See [../../architecture.md](../../architecture.md) and the ADRs in
`docs/ADR/`. The system is a modular monolith with core + compiled-in plugins.

## Dependencies

- Go 1.27+
- PostgreSQL
- pgx, sqlc, Goose, Task
- Docker (optional)

## Risks

- Over-engineering the plugin system — mitigated by keeping the contract
  minimal and adding no speculative abstractions.
- Schema drift — mitigated by sqlc using the migration files as schema source.

## Tickets

1. [001-project-skeleton-and-toolchain.md](ticket/001-project-skeleton-and-toolchain.md)
2. [002-core-registries-and-event-bus.md](ticket/002-core-registries-and-event-bus.md)
3. [003-http-api-and-readiness.md](ticket/003-http-api-and-readiness.md)
4. [004-postgresql-migrations-and-sqlc.md](ticket/004-postgresql-migrations-and-sqlc.md)
5. [005-azeroth-soap-adapter.md](ticket/005-azeroth-soap-adapter.md)
6. [006-plugin-system-and-initial-plugins.md](ticket/006-plugin-system-and-initial-plugins.md)
