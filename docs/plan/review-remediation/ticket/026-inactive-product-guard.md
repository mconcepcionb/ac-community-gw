# 026 — Inactive product purchase guard

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 007

## Goal

Prevent purchasing a product that has been deactivated.

## Findings addressed

- Medium: `handlePurchase` fetches via `ProductBySKU` (which does not filter on
  `active`) and never checks `product.Active`; the catalog hides inactive
  products, but a client who knows the SKU can still buy one.

## Context

`internal/plugins/azerothstore/handlers.go`. `SetProductActive` already exists.

## Atomic change

Reject purchases of inactive products in the purchase handler (and any other
buy/order path), returning `404 product_not_found` or `409 product_inactive`
consistently.

## Requirements

- In `handlePurchase`, after fetching the product, return the chosen error when
  `!product.Active`.
- Choose and document the disclosure policy: prefer `404 product_not_found` so
  inactive products are indistinguishable from missing ones.
- Verify no other path (e.g. direct order creation) bypasses the check.
- Keep the admin endpoints able to read/update inactive products.

## Tests

- Test: purchasing an active product succeeds.
- Test: purchasing an inactive product returns the chosen error and does not
  debit the wallet.
- Test: an admin can still fetch/update an inactive product.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green.
- No purchase path can debit points for an inactive product.

## Rollback

Revert; inactive products become purchasable again.

## Out of scope

- Soft-delete semantics.
- Refunds for existing inactive-product orders.
