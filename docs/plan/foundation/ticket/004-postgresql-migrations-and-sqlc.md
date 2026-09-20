# PostgreSQL, migrations and sqlc

## Goal

Persist the gateway's own state in PostgreSQL with explicit SQL, Goose
migrations and typed sqlc access.

## Context

The gateway database is PostgreSQL, distinct from AzerothCore's MySQL/MariaDB
databases. See ADR 0003.

## Requirements

- `database/sql` with the pgx driver.
- Goose SQL migrations in `migrations/` with a single global sequence.
- Migrations embedded in the binary for optional startup application.
- sqlc configuration using `migrations/` as the schema source.
- Initially owned tables: `roles`, `permissions`, `role_permissions`,
  `audit_log`, `community_users`, `discord_identities`, `sessions`,
  `azeroth_account_links`.
- Generated code committed.

## Acceptance criteria

- `task db:migrate` applies all migrations.
- `task db:status` reports the applied versions.
- `task codegen:sqlc` regenerates without diff.
- The server can connect and `readyz` reports PostgreSQL availability.

## Implementation notes

`migrations/embed.go` exposes the SQL files via `embed.FS` so both the Goose CLI
and the server use one source of truth. Query ownership is per module.

## Tests

- `sqlc generate` runs clean in CI (`task codegen:sqlc`).
- Manual/integration: migrations up and down against a disposable database.
