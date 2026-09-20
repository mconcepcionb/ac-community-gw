# 008 — Order lifecycle idempotency and integrity

**Phase:** 2 · **Gate:** G-standard, contributes to G2-correctness · **Depends on:** 007

## Goal

Make order completion/failure terminal and idempotent so a wallet is never
double-refunded and a delivered order can never be refunded.

## Findings addressed

- High: `UpdateOrderStatus` has no state predicate and `FailOrder` credits
  unconditionally; duplicate/concurrent `FailOrder` calls double-refund and a
  `FailOrder` after delivery also refunds.
- Medium: `store_orders.status` has no `CHECK`; `price_points` has no
  non-negative check.
- Low: `FailOrder` does not call `EnsureWallet`.
- Low: `handlePurchase` can fail `CompleteOrder` after a successful delivery,
  leaving the order `pending` with points debited (mitigated by 027, but the
  state machine must be enforced here first).

## Context

`internal/plugins/azerothstore/repository/queries.sql` and `store.go`, plus
`migrations/00005_store.sql`. The order lifecycle is `pending → delivered |
failed`; refunds happen only when moving to `failed`.

## Atomic change

Enforce the state machine in SQL, add the missing constraints, and make refunds
happen exactly once. This ticket adds a migration.

## Requirements

- `UpdateOrderStatus` gains `AND status = 'pending'` and returns no row when the
  order is already terminal; the repository maps that to a new
  `domain.ErrOrderNotPending`.
- `FailOrder`: call `EnsureWallet` first; only credit when the transition
  succeeded; never credit an order that was already delivered/failed.
- Add migration `00006_order_integrity.sql`:
  - `ALTER TABLE store_orders ADD CONSTRAINT store_orders_status_check
    CHECK (status IN ('pending','delivered','failed'))`.
  - `ALTER TABLE store_orders ADD CONSTRAINT store_orders_price_points_check
    CHECK (price_points >= 0)`.
- `CompleteOrder` sets `status = 'delivered'` only from `pending` (same guard);
  if the order is already delivered, treat as an idempotent no-op success.
- Return typed errors so handlers can respond `409 order_not_pending` instead of
  a generic 503.
- Down migration drops the constraints.

## Tests

- Repository integration test: `FailOrder` twice credits the wallet once; the
  second call returns `ErrOrderNotPending`.
- Test: `FailOrder` on a delivered order does not credit.
- Test: `FailOrder` on a user without a wallet still refunds (EnsureWallet).
- Test: `CompleteOrder` after delivered is a no-op success.
- Migration test: `up`/`down` on a scratch database; duplicate/negative status
  rows are rejected by the constraints.

## Acceptance criteria (gate G-standard, G2-correctness)

- `task check` green; integration tests pass against Postgres.
- No code path can debit once and refund twice.
- The status constraint exists and is exercised.

## Rollback

Revert the migration (`down`) and the repository changes. Existing terminal
orders are unaffected.

## Out of scope

- Reconciliation of a delivered-but-not-completed order (ticket 027).
- Wallet ledger FKs/indexes (ticket 020).
