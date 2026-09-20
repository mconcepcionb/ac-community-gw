# Store: wallet grant

## Goal

Type and document `POST /api/v1/store/wallets/grant`.

## Context

Administrative wallet top-up used by staff, distinct from purchases.

## Requirements

- Typed request/response DTOs, preserving fields.
- swag annotations under `@Tags store`: `@ID store.wallets.grant`,
  `@Accept json`, `@Success 200`, `@Failure 400/403/404/422/503` and the shared
  envelope; required permission documented.

## Acceptance criteria

- Exact-JSON test for success and for a malformed body.
- Operation present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
