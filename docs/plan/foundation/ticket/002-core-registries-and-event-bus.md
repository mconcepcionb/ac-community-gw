# Core registries and event bus

## Goal

Provide the transversal registries that allow plugins to cooperate without
importing each other.

## Context

Queries, commands, events and permissions are distinct interaction kinds and
must not be conflated. See ADR 0009 and ADR 0010.

## Requirements

- Command registry: `Register` / `Execute`, duplicate and not-found errors,
  typed helpers.
- Service/capability registry: `Publish` / `Resolve`, typed helpers.
- Event bus: synchronous, in-process, non-durable, ordered handlers, joined
  errors, no automatic goroutines.
- Permission registry + in-memory `Authorizer` (role -> permissions).
- Plugin contract and manager.

## Acceptance criteria

- Registries are safe for concurrent use.
- Duplicate registration is an error.
- Event `Publish` runs all handlers even if one fails and returns joined errors.
- No global mutable state; registries are constructed in `cmd/server`.

## Implementation notes

Keep implementations small. Do not turn the command registry into a CQRS
framework or the service registry into an opaque service locator.

## Tests

- Command registry: register, execute, duplicate, not found, invalid payload,
  sorted names.
- Service registry: publish, resolve, duplicate, not found, type mismatch.
- Event bus: order, joined errors, typed subscribe, unrelated event, empty name.
- Permissions: register, duplicate, empty name, role grants.
