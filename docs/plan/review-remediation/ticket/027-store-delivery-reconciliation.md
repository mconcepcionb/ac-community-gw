# 027 — Store delivery reconciliation

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 008

## Goal

Guarantee that an order always reaches a terminal state and that points are
never left debited for a successful delivery.

## Findings addressed

- Low/Medium: in `handlePurchase`, if `CompleteOrder` fails after `Deliver`
  succeeded, the handler returns `503` but the wallet debit stands and the order
  stays `pending`; there is no reconciliation or refund path.

## Context

`internal/plugins/azerothstore/handlers.go`, `repository/store.go`. Ticket 008
made the state machine terminal and idempotent; this ticket handles the gap
between delivery success and state persistence.

## Atomic change

Introduce a bounded best-effort completion with a fallback that never loses the
delivery output, plus a reconciliation mechanism for orders left `pending`.

## Requirements

- On `CompleteOrder` failure after a successful delivery: retry a bounded number
  of times; if still failing, persist the delivery output by an alternate route
  (e.g. write the output into the order via a direct update) or record a
  reconciliation marker.
- Never refund a delivered order (ticket 008 guarantees this in SQL); instead,
  move it to `delivered` on reconciliation.
- Add a reconciliation task/command that scans `pending` orders with a known
  delivery output and completes them, and reports `pending` orders with no output.
- Expose the reconciliation outcome via logs/metrics.
- Document the runbook step in `docs/runbooks/` (or `docs/store.md`).

## Tests

- Test: injecting a `CompleteOrder` failure keeps the order `pending` but
  records the delivery output so reconciliation can finish it.
- Test: the reconciliation task completes such an order without a second debit
  or refund.
- Test: reconciliation never refunds a delivered order.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green; integration tests pass.
- No code path can leave points debited with no route to a terminal state.

## Rollback

Revert; the failure window returns.

## Out of scope

- Distributed transactions with the game server.
- Refunding genuinely failed deliveries (already handled by ticket 008).
