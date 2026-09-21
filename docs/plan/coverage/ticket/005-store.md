# store

**Milestone:** B · **Gate:** G-standard · **Depends on:** 001, 003

## Goal

Raise the `store` plugin from 38.2% to ≥ 60%, covering the order lifecycle and
the untested handlers.

## Context

`internal/plugins/store` is the gateway store plugin. Tested today: purchase,
insufficient funds, missing link, inactive product, completion-failure output,
grant and admin orders (`handlers_test.go`), plus product catalog
(`products_test.go`). Untested main files: `resolve.go` (refund/retry),
`account.go` (wallet/orders service), `handleProduct`, `handleOrders`,
`handleUpdateProduct`, product CRUD, `Register` and `permissionDefs`
(`internal/plugins/store/{resolve.go,account.go,handlers.go}`).

## Requirements

- `resolve.go`: refund and retry — success, wrong state (`409`
  `order_not_pending`), reason required, audit recorded, and refund idempotency
  (no double credit; the SQL state machine is the guard).
- `account.go`: wallet and order reads, totals.
- `handlers.go`: `handleProduct`, `handleOrders`, `handleUpdateProduct`, product
  create/update/set-active paths.
- `Register` and `permissionDefs`.
- Keep the existing fakes (`fakeStore`, `fakeDelivery`, `fakeAccounts`,
  `fakeUsers`, `purchasePlugin`, `authedRequest`) and move shared request code to
  `internal/testsupport`.

## Tests

- Lifecycle test proving an order cannot be refunded twice and a delivered order
  cannot be refunded.
- Refund credits exactly once and completes the order.
- Retry transitions only from `pending`.

## Acceptance criteria

- `store` reaches ≥ 60%.
- Every state-machine guard has a negative test.
- `task check` green.

## Out of scope

- The store repository (C9, integration).
- Schema changes.
