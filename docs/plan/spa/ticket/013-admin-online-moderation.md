# Admin: online and moderation

## Goal

Type and document the live server and player moderation endpoints.

## Context

`azeroth-admin` serves `GET /api/v1/azeroth/online`,
`POST /api/v1/azeroth/players/{name}/kick|mute|unmute` and
`POST /api/v1/azeroth/characters/{name}/ban|unban`.

## Requirements

- Typed request/response DTOs for all operations, preserving fields.
- `online` list is always an array.
- swag annotations under `@Tags azeroth-admin` with stable `@ID`s:
  `azeroth.online.list`, `azeroth.players.kick|mute|unmute`,
  `azeroth.characters.ban|unban`.
- `@Success`/`@Failure` per operation plus the shared envelope; required
  permission documented.

## Acceptance criteria

- Exact-JSON tests for the online list and each action (success + upstream
  failure).
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
