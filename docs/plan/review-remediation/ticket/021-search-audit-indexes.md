# 021 — Search and audit indexes

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 020

## Goal

Make user search and audit queries efficient and consistent, and give the audit
table a retention/partitioning story.

## Findings addressed

- Medium: the user-search query uses leading-wildcard `ILIKE` across several
  columns and a case-sensitive `LIKE` on `discord_id`; there is no supporting
  index, forcing sequential scans, and the mixed semantics are surprising.
- Medium/Low: `audit_log` grows unbounded with only `occurred_at` indexed; no
  retention or partitioning; `LIKE` on `discord_id` treats `%`/`_` as wildcards.

## Context

`internal/plugins/identitydiscord/repository/queries.sql`, `migrations/00001_core.sql`.
This ticket adds a migration and depends on 020.

## Atomic change

Add `00008_search_audit_indexes.sql`, enable `pg_trgm`, add GIN indexes for the
searchable columns, extend audit indexes, and normalize the search query to a
consistent predicate.

## Requirements

- Enable the `pg_trgm` extension (or explicitly document the decision not to and
  accept sequential scans at the expected data size).
- Add GIN trigram indexes for the columns searched:
  `community_users.display_name`, `discord_identities.username`,
  `discord_identities.global_name`.
- Use `ILIKE` consistently and escape `%`/`_` in the filter before building the
  pattern.
- Add indexes for common audit filters: `actor_id`, `action`, `target_type`.
- Define a retention/partitioning approach for `audit_log` (partition by month or
  a documented cleanup job/task); do not delete history implicitly in this
  ticket, only enable the mechanism.
- Down migration drops what it added (extension drop is optional/documented).

## Tests

- Query-plan test (or integration assertion) that the search uses the index for a
  representative term.
- Test: a filter containing `%`/`_` does not match unintended rows.
- Migration `up`/`down` test.
- If partitioning is chosen, a test that inserts across partition boundaries and
  queries back.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green; integration tests pass.
- Search uses an index for a term of length ≥ 3 (or the deviation is documented).
- Audit filters are indexed.

## Rollback

Run the down migration.

## Out of scope

- Rewriting the search UX (frontend tickets).
- Full-text search.
