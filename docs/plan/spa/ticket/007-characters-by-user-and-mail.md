# Characters: by user and mail

## Goal

Type and document the per-user character endpoint and the mail delivery
endpoint.

## Context

`azeroth-character` serves `GET /api/v1/azeroth/users/{user_id}/characters` and
`POST /api/v1/azeroth/mail` (`mail.go`).

## Requirements

- Typed DTOs for both responses and for the mail request body.
- Preserve current JSON keys, types and validation behaviour (unknown-field
  rejection, body size limit).
- swag annotations under `@Tags azeroth-character`:
  - `@ID azeroth.user_characters.list` (path param `user_id`),
  - `@ID azeroth.mail.send` (`@Accept json`).
- `@Success`, `@Failure 400/403/404/502/503` and the shared envelope;
  required permission documented in each description.

## Acceptance criteria

- Exact-JSON tests for both endpoints; malformed mail body still returns
  `400 bad_request`.
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001. Independent from 006 but shares the plugin's DTO file.
