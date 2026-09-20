# Admin: bans and gmlevel

## Goal

Type and document the account moderation endpoints for bans and GM level.

## Context

`azeroth-admin` serves `POST /api/v1/azeroth/accounts/{username}/ban`,
`POST /api/v1/azeroth/accounts/{username}/unban` and
`PUT /api/v1/azeroth/accounts/{username}/gmlevel`.

## Requirements

- Typed request/response DTOs for the three endpoints.
- swag annotations under `@Tags azeroth-admin`:
  `@ID azeroth.accounts.ban|unban|set_gmlevel`; path param `username`.
- `@Accept json`, `@Success` codes equal to today, failure cases and the shared
  envelope; required permission documented per operation.

## Acceptance criteria

- Exact-JSON tests for success and representative failures (bad body, upstream
  error).
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001. Can land together with 013 if it stays reviewable.
