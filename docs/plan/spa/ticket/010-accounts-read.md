# Accounts: read

## Goal

Type and document `GET /api/v1/azeroth/accounts`.

## Context

Account listing reads the AzerothCore login database through
`azerothdb.AccountReader`; it can be unavailable when the DB is not configured.

## Requirements

- Typed DTO for the account list response (fields exactly as today).
- `accounts` is always an array.
- swag annotations under `@Tags azeroth-account`: `@ID azeroth.accounts.list`,
  query params `filter`, `limit`, `offset`; `@Success 200`, `@Failure 503` and
  the shared envelope; required permission documented.

## Acceptance criteria

- Exact-JSON test with a fake reader and a test for the unconfigured case.
- Operation present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001. Same DTO file as 009 but a separate change.
