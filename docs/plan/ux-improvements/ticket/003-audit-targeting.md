# Audit targeting and metadata

## Goal

Make the audit log usable as a per-entity history panel by filtering on the
target and exposing the stored metadata.

## Context

`GET /api/v1/admin/audit` (permission `gw.audit.read`, ADR 0014) supports
`actor`, `target` (substring only), `action`,
`since`, `until`, `limit`, `offset` (`internal/plugins/azerothadmin/auditview.go:52-80`).

Problems found:

- No `target_type` filter, even though the column exists and is indexed
  (`migrations/00008_search_audit_indexes.sql:17`).
- `target_id` values are heterogeneous (account usernames, character names,
  `role:permission`, Discord role ids, report ids), so a substring match is
  ambiguous.
- `metadata jsonb` is written/read by the store but never returned in the DTO
  (`auditview.go:16-26`).

## Requirements

- Add `target_type` to `audit.ListFilter` and the Postgres query, and to the
  `GET /admin/audit` query string.
- Expose `metadata` on `AuditEntry` (omit when empty).
- Ensure account and character actions record a stable `target_type`
  (`account`, `character`, `user`, `role`, `discord-role`, `report`).
- Add an index on `(target_type, target_id, occurred_at DESC)` if the existing
  indexes do not cover the panel query.

## Acceptance criteria

- `GET /admin/audit?target_type=account&target=<username>` returns only that
  account's history, newest first.
- Metadata round-trips and is visible in the API response.
- Account/character mutations set their `target_type` consistently.
- `task check` and `task openapi:check` green.

## Implementation notes

- Keep the substring `target` filter as a convenience search; add the exact
  `(target_type, target_id)` pair as the panel query.
- Consider a dedicated `?target_type=&target_id=` (exact) pair distinct from the
  fuzzy `target`; document the difference.

## Tests

- Go tests for the new filter and metadata serialization.
- Update `admin-audit-page.test.tsx` if the filters broaden.

## Dependencies

- Independent; consumed by 005 and 008 for per-entity history tabs.
