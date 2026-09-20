# Unified moderation and reconciliation queue

## Goal

Give staff a single queue of items awaiting action so nothing is dropped.

## Context

This ticket covers use case G5 ([../../../use-cases.md](../../../use-cases.md)).
Order reconciliation is implicit in the store state machine today
([../../../store.md](../../../store.md)).

## Requirements

- Backend: expose pending reconciliation items (store), failed deliveries,
  pending account claims (011) and recent bans/mutes through one query surface.
- Console `/admin/moderation`: a queue with type filters, age and assignee, and a
  one-action resolution where safe.
- Resolutions are idempotent and audited; resolving twice has no second effect.
- Reuse the store reconciliation path and its state-machine guarantees.

## Acceptance criteria

- Each item type is listed with enough context and, where safe, a one-click
  resolution.
- Double resolution is a no-op; the store is never double-refunded.
- The queue is visible only to permitted staff.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Keep the queue read-mostly; actions delegate to existing commands.
- The reconciliation action must not bypass the `pending -> delivered | failed`
  transition rules.

## Tests

- Go tests for idempotent resolution; RTL + MSW for the queue and filters.

## Dependencies

- 002, 004. 011 contributes pending claims.
