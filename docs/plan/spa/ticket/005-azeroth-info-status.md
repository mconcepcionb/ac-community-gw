# Azeroth info: status

## Goal

Type and document `GET /api/v1/azeroth/info/status`.

## Context

`azeroth-info` reports AzerothCore server status over the command executor and
parses the output in `parse.go`.

## Requirements

- Introduce typed request/response DTOs for the endpoint (whatever the parser
  currently yields), preserving field names and types exactly.
- swag annotations: `@Tags azeroth-info`, `@ID azeroth.info.status`,
  `@Success 200`, `@Failure 502`, `@Failure 503`, shared error envelope.
- Document the required permission in the description.

## Acceptance criteria

- JSON on the wire is unchanged (exact-JSON test with a fake executor).
- Operation `azeroth.info.status` present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
