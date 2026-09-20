# ADR 0006: Module-owned AzerothCore commands

## Status

Accepted

## Context

AzerothCore is driven through CLI commands sent over a small SOAP
`executeCommand` interface. These command strings are an internal detail that
must not leak to API consumers, and they must not become a global, shared
catalogue maintained in the core or in the SOAP adapter.

## Decision

The plugin that owns a capability is also the owner of the AzerothCore CLI
syntax required to implement it.

- The SOAP adapter knows only transport.
- The core exposes an application command registry with typed payloads.
- Each plugin registers application commands and builds its own AzerothCore
  command strings internally.

## Consequences

- `.send items ...` is known only to the plugin that owns that capability.
- Changing AzerothCore syntax is a local change within one plugin.
- No global package of AzerothCore commands exists.
- There is no generic `execute-command` API, by construction.
