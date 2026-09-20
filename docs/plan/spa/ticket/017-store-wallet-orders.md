# Store: wallet and orders

## Goal

Type and document the wallet and order endpoints.

## Context

`azeroth-store` serves `GET /api/v1/store/wallet`,
`GET /api/v1/store/orders` and `POST /api/v1/store/orders`.

## Requirements

- Typed DTOs for the wallet, the order list and the create-order request.
- List responses are always arrays.
- swag annotations under `@Tags store`: `@ID store.wallet.get`,
  `store.orders.list`, `store.orders.create`; `@Accept json`,
  `@Success`/`@Failure` per operation and the shared envelope; required
  permission documented.

## Acceptance criteria

- Exact-JSON tests for the three operations.
- Operations present in `api/swagger.yaml`.
- `task check` green.

## Dependencies

- 001.
