# ADR 0001: Modular monolith

## Status

Accepted

## Context

The gateway bridges several distinct domains (identity, accounts, administration,
store, later characters and guilds) and integrates two very different external
systems (Discord and AzerothCore). We need clear module boundaries without the
operational cost of distributed systems.

## Decision

Build a **modular monolith**. All plugins are statically compiled into a single
binary. There are no `.so` plugins, no runtime plugin loading, no RPC between
processes, no microservices and no service mesh.

Modules are separated by package boundaries and enforced import rules, and they
communicate through core registries.

## Consequences

- One build, one deployable, one process to operate.
- Module boundaries must be enforced by discipline and tests, since the
  compiler does not enforce "no cross-plugin imports" by itself; this is covered
  by `internal/architecture`.
- Scaling is vertical for now; this is acceptable at the expected scale.
