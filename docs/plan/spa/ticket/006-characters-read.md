# Characters: read

## Goal

Type and document the character read endpoints.

## Context

`azeroth-character` serves `GET /api/v1/azeroth/characters` and
`GET /api/v1/azeroth/characters/{name}` from `characters.go`.

## Requirements

- Typed DTOs for the list and the detail responses, preserving every field.
- Array fields are always `[]`, never `null`.
- swag annotations under `@Tags azeroth-character`:
  - `@ID azeroth.characters.list` (query params `filter`, `limit`, `offset`),
  - `@ID azeroth.characters.get` (path param `name`).
- `@Success 200`, `@Failure 404`, `@Failure 503` and the shared envelope; the
  required permission is documented in each description.

## Acceptance criteria

- Exact-JSON tests for list and detail against fake readers.
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
