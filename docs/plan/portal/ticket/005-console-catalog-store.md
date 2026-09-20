# Console catalog and store operations

## Goal

Consolidate item browsing, catalog management, wallet operations and order
monitoring in the console.

## Context

Today `/items`, the admin view of `/store/products` and `/store/wallet` are
separate. This ticket covers use cases S1, S2 and S3
([../../../use-cases.md](../../../use-cases.md)) and
[../../../store.md](../../../store.md).

## Requirements

- `/admin/items`: item catalog search and detail (moved from `/items`).
- `/admin/store`: product list and CRUD (moved from the admin `/store/products`).
- `/admin/store/wallets`: wallet lookup and points grant (C2/S2).
- `/admin/store/orders`: order list with status filter and refund visibility
  (S3); manual resolution arrives in 015.
- Admin order reads are separate from the self-scoped player permission
  (`store.orders.read`).

## Acceptance criteria

- Catalog CRUD validates item ids and renders items like the storefront (010).
- Orders show failed orders and whether points were refunded.
- The old `/items` path and the admin `/store/products` path are removed.
- `task web:check` green.

## Implementation notes

- Keep the existing product form and item render components; only rehome them.
- Share the item render component with the portal store (010).
- Add or reuse an admin-scoped orders read if the existing one is self-scoped.

## Tests

- Move the existing store and items tests; extend for the order filters.

## Dependencies

- 001, 002. Extended by 015 (queue); shares the render with 010.
