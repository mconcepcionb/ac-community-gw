# Portal store

## Goal

Deliver the player storefront, purchase flow, wallet and order history.

## Context

The store endpoints already exist; the current `/store/products` page is
admin-oriented. This ticket covers use cases P6, P7 and P8
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- `/store` and `/store/products/$sku`: player storefront with the same item
  rendering as the admin preview; inactive products hidden.
- Purchase via `POST /api/v1/store/orders`, requiring a linked account and an
  owned character, with clear insufficient-funds and delivery-outcome messaging.
- `/wallet`: points balance, append-only ledger and orders with
  pending/delivered/failed status.
- Gate with `store.catalog.read`, `store.purchase`, `store.wallet.read` and
  `store.orders.read`.

## Acceptance criteria

- A player can browse, buy and watch the order progress to delivered.
- Insufficient funds debit nothing; a delivery failure refunds and is shown as
  failed.
- A player cannot see another user's wallet.
- `task web:check` green.

## Implementation notes

- Share the item render component with the console store (004).
- The purchase dialog exists; adapt it to the portal and the onboarding
  prerequisite.

## Tests

- MSW for purchase success, insufficient funds, delivery failure and wallet
  isolation.

## Dependencies

- 001. 007 provides owned-character selection; 010/011 handle unlinked users.
