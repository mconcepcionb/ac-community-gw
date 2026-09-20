# Admin: announce

## Goal

Type and document `POST /api/v1/azeroth/announce`.

## Context

`azeroth-admin` broadcasts a server-wide announcement through the command
executor.

## Requirements

- Typed request/response DTOs.
- swag annotations under `@Tags azeroth-admin`: `@ID azeroth.announce.send`,
  `@Accept json`, `@Success 200`, failure cases and the shared envelope;
  required permission documented.

## Acceptance criteria

- Exact-JSON test for success and for a malformed body.
- Operation present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
