# ADR 0003: PostgreSQL with sqlc and Goose

## Status

Accepted

## Context

The gateway needs its own transactional store, separate from AzerothCore's
MySQL/MariaDB databases. We want explicit SQL, static typing and simple
operations, without the opacity of an ORM.

## Decision

- Database: PostgreSQL.
- Access: `database/sql` with the pgx driver.
- Code generation: sqlc, using the Goose migrations as the schema source.
- Migrations: Goose, SQL-first, single global sequence in `migrations/`.

## Consequences

- SQL is written and reviewed explicitly; queries are typed at compile time.
- The schema has one source of truth; sqlc does not require a duplicated schema.
- Generated code is committed to the repository.
- Developers must run `task codegen:sqlc` after changing queries or migrations.
- No ORM: some boilerplate is expected and accepted.
