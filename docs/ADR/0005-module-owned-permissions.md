# ADR 0005: Module-owned permissions

## Status

Accepted

## Context

Authorization must be consistent and free of scattered role checks such as
`if user.IsAdmin`. At the same time, the core must not grow a global catalogue
of every permission in the system.

## Decision

Each plugin owns and registers its own permissions. The core provides only the
permission registry, role -> permission mappings, checks and middleware.
Authorization is expressed as permissions, never as role or admin special-cases.

## Consequences

- Adding a permission never requires changing the core.
- Permission names are namespaced by domain (`azeroth.account.manage`,
  `identity.self.read`).
- The core stays domain-agnostic.
- Duplicate or conflicting permission names are detectable at registration time.
