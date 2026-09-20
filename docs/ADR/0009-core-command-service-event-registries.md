# ADR 0009: Core command, service and event registries

## Status

Accepted

## Context

Plugins need to cooperate without importing each other. Queries, commands,
events and permissions are fundamentally different interactions and should not
be conflated.

## Decision

The core provides four small registries:

- **Command registry** — request an action (`account.ban`,
  `character.send-item`), typed payloads, single owner per command.
- **Service / capability registry** — synchronous reads and validations
  (`CharacterReader.OwnedBy`), consumed via consumer-declared interfaces.
- **Event bus** — announcements of facts that already happened.
- **Permission registry** — authorization definitions and checks.

The registries are deliberately minimal. The command registry is not a CQRS
framework; the service registry is not an opaque service locator.

## Consequences

- Plugins depend on core contracts only.
- Events are not used as commands or queries.
- Commands are not used as a generic query mechanism.
- The core stays small and free of domain knowledge.
