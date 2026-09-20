# Audit log viewer

## Goal

Let staff search the audit log and let auditors read it without mutating.

## Context

Audit is written today but not surfaced in the SPA. This ticket covers use case
A1 and supports G6 and O5 ([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- Backend: an audit read capability and a read permission, with paginated,
  filterable queries by actor, target, action and time range.
- Console `/admin/audit`: a search UI, read-only, linked from the 360 view (003).
- A read-only staff role can open the console and this page without action
  controls (O5).
- Entries never contain secrets, tokens or passwords.

## Acceptance criteria

- Filtering by actor, target, action and time works and paginates.
- The viewer is read-only; no edit or delete affordances exist.
- A read-only role sees the audit page but cannot perform actions elsewhere.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Add the audit read permission to the registry and grant it to an auditor role
  (role administration proper lands in 014).
- Redaction is enforced server-side; the SPA renders what it receives.

## Tests

- Go tests for filtering and redaction; RTL + MSW for the viewer and role gating.

## Dependencies

- 001, 003.
