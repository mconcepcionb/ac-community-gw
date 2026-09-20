# Account links

## Goal

Type and document the community-user to AzerothCore-account link CRUD.

## Context

`azeroth-account` serves:
`POST /api/v1/azeroth/account-links`,
`GET /api/v1/azeroth/account-links`,
`GET /api/v1/azeroth/account-links/{user_id}`,
`DELETE /api/v1/azeroth/account-links/{user_id}` (`links.go`).

## Requirements

- Typed request/response DTOs for all four operations, preserving fields.
- List responses are arrays, never `null`.
- swag annotations under `@Tags azeroth-account`:
  `@ID azeroth.account_links.create|list|get|delete`; path param `user_id`.
- `@Success`/`@Failure` per operation plus the shared envelope; required
  permission documented.

## Acceptance criteria

- Exact-JSON tests for each operation.
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
