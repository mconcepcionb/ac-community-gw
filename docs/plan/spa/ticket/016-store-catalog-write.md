# Store: catalog write

## Goal

Type and document the store product write endpoints.

## Context

`azeroth-store` serves `POST /api/v1/store/products`,
`PUT /api/v1/store/products/{sku}` and `DELETE /api/v1/store/products/{sku}`.

## Requirements

- Typed request/response DTOs for the three operations, preserving fields and
  validation.
- swag annotations under `@Tags store`: `@ID store.products.create|update|delete`;
  `@Accept json`, `@Success` codes equal to today, `@Failure 400/403/404/409/422/503`
  and the shared envelope; required permission documented.

## Acceptance criteria

- Exact-JSON tests for success and representative failures.
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
