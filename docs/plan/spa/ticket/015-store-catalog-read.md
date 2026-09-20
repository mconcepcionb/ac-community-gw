# Store: catalog read

## Goal

Type and document the store product read endpoints.

## Context

`azeroth-store` serves `GET /api/v1/store/products` and
`GET /api/v1/store/products/{sku}` (`handlers.go`).

## Requirements

- Typed DTOs for the product list and detail, preserving fields.
- `products` is always an array.
- swag annotations under `@Tags store`: `@ID store.products.list` (query params
  `filter`, `limit`, `offset`) and `@ID store.products.get` (path param `sku`);
  `@Success 200`, `@Failure 404`, `@Failure 503` and the shared envelope;
  required permission documented.

## Acceptance criteria

- Exact-JSON tests for list and detail.
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
