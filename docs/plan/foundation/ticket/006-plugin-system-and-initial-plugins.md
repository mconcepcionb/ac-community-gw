# Plugin system and initial plugins

## Goal

Provide the minimal plugin contract and the initial compiled-in plugin packages.

## Context

The modular monolith needs a small, explicit plugin interface and a registry of
core facilities passed during registration. See ADR 0001 and ADR 0008.

## Requirements

- `Plugin` interface: `Name()` and `Register(ctx, *Registry)`.
- Manager registering plugins in order, rejecting duplicate names.
- Registry exposing mux, registries, audit and auth middleware.
- Initial plugins: `identity-discord`, `azeroth-account`, `azeroth-admin`,
  `azeroth-info`, `azeroth-store`.
- Each plugin registers its own permissions; account/admin register commands;
  info publishes a capability and mounts a protected route.

## Acceptance criteria

- All plugins register without error in `cmd/server`.
- Registered commands, services and permissions are logged at startup.
- A plugin can only reach other domains through the registries.
- Architectural import boundaries hold.

## Implementation notes

Plugins receive concrete dependencies (session manager, command executor) from
their constructors in `cmd/server`. Registration only wires registries.

## Tests

- `internal/architecture` enforces no cross-plugin, no core->plugin and
  no plugin->adapter imports.
- Registry and bus tests already cover the cooperation mechanisms.
