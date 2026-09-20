# Items: read

## Goal

Type and document the item catalog read endpoints.

## Context

`azeroth-item` serves `GET /api/v1/azeroth/items` and
`GET /api/v1/azeroth/items/{entry}` from `handlers.go`, returning
`map[string]any{"items": []itemview.View}` and a single view.

## Requirements

- Typed DTOs wrapping `itemview.View` so the JSON matches today exactly.
- `items` is always an array.
- swag annotations under `@Tags azeroth-item`:
  - `@ID azeroth.items.list` (query params `filter`, `limit`, `offset`, `class`),
  - `@ID azeroth.items.get` (path param `entry`).
- `@Success 200`, `@Failure 404`, `@Failure 422`, `@Failure 503` and the shared
  envelope; required permission documented.

## Acceptance criteria

- Exact-JSON tests for list and detail; invalid `class`/`entry` keep returning
  the current errors.
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
