# Accounts: write

## Goal

Type and document the account write endpoints.

## Context

`azeroth-account` serves `POST /api/v1/azeroth/accounts`,
`POST /api/v1/azeroth/accounts/{username}/password` and
`PUT /api/v1/azeroth/accounts/{username}/email` (`http.go`).

## Requirements

- Typed request/response DTOs for the three endpoints, preserving field names
  and validation.
- swag annotations under `@Tags azeroth-account`:
  - `@ID azeroth.accounts.create`,
  - `@ID azeroth.accounts.set_password`,
  - `@ID azeroth.accounts.set_email`.
- `@Accept json`, `@Success` codes equal to today, `@Failure 400/403/409/422/502/503`
  and the shared envelope; required permission documented per operation.

## Acceptance criteria

- Exact-JSON tests for success and the main failure paths (bad body -> 400,
  conflict -> 409, upstream failure -> 502).
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
