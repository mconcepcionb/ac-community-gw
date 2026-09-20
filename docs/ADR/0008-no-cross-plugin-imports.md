# ADR 0008: No cross-plugin imports

## Status

Accepted

## Context

In a modular monolith, package imports are the main source of accidental
coupling. A plugin importing another plugin's implementation would create
circular dependencies, make ownership ambiguous and defeat the module
boundaries.

## Decision

Enforce these import rules:

```
core MUST NOT import plugins
plugins MAY import core contracts
a plugin MUST NOT import another plugin's implementation
core and plugins MUST NOT import adapters
```

Cross-plugin interaction happens exclusively at runtime through core registries
and Go interfaces (structural typing), not through package imports.

## Consequences

- Modules remain independently understandable and replaceable.
- Cross-plugin contracts are declared by the consumer, so no shared "contracts"
  package is needed.
- The rules are enforced by tests (`internal/architecture`), failing the build
  on violation.
- Importing a shared adapter directly is forbidden, keeping adapters
  replaceable and injected from `cmd/server`.
