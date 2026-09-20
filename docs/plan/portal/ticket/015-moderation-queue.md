# Unified moderation and reconciliation queue

## Goal

Give staff a single, filterable queue of items awaiting action so nothing is
dropped, including the ability to resolve stuck store orders.

## Context

This ticket covers use case G5 ([../../../use-cases.md](../../../use-cases.md)).
The periodic reconciler (`cmd/server/main.go`) completes pending orders whose
delivery output was persisted after a successful delivery; the genuinely stuck
case is a **pending order with no delivery output**, which only a human can
resolve.

## Requirements

- Backend: expose one query surface for:
  - stuck store orders (pending, no delivery output);
  - auto-reconciled history (orders the reconciler completed);
  - failed/refunded orders;
  - pending and failed account claims (012);
  - recent bans, mutes and kicks;
  - player reports (014).
- Console `/admin/moderation`: a filterable list sorted by age. No assignment or
  SLA workflow in this iteration.
- On a stuck order, staff may **refund** or **retry delivery then complete**.
  Each action requires confirmation and a reason and is audited.
- Resolutions are idempotent: the SQL state machine
  (`pending -> delivered | failed`) prevents a double refund or completing a
  delivered order.

## Acceptance criteria

- Each item type is listed with enough context and, where safe, a one-click
  resolution.
- Refunding or retrying a stuck order is confirmed, reasoned and audited; a
  second attempt on the resolved order is a no-op.
- A delivered order can never be refunded; a refunded order can never be
  delivered.
- The queue is visible only to permitted staff.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Actions delegate to store commands; the queue never mutates order state
  directly.
- Auto-reconciled and failed/refunded entries are read-only history.

## Tests

- Go tests for idempotent resolution, refund-vs-delivered guards and reason
  capture; RTL + MSW for the queue and filters.

## Dependencies

- 003, 005, 012, 014.
