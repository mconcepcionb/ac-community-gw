# Documentation

This directory separates documentation by purpose:

| Location | Describes |
| --- | --- |
| `docs/*.md` | How the system works **today** |
| `docs/ADR/` | **Why** important decisions were taken |
| `docs/plan/` | **What** we intend to build and how |
| `docs/runbooks/` | **How to operate** the system |

Stable documents are not a backlog. When a plan is delivered, its knowledge is
consolidated here and its decisions into an ADR.

## Stable documents

- [architecture.md](architecture.md) — modular monolith, dependency rules
- [plugins.md](plugins.md) — plugin contract and lifecycle
- [authentication.md](authentication.md) — Discord identity and sessions
- [permissions.md](permissions.md) — RBAC and permission ownership
- [store.md](store.md) — points, catalog, orders and in-game delivery
- [security.md](security.md) — trust boundaries and handling of secrets
- [database.md](database.md) — PostgreSQL ownership, migrations, sqlc
- [azerothcore-integration.md](azerothcore-integration.md) — SOAP adapter and command ownership
- [frontend.md](frontend.md) — SPA structure, contract client, auth flow, testing
- [development.md](development.md) — toolchain and workflow
