# 023 — Domain events: publish or remove

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 001

## Goal

Resolve the dead domain-event definitions: either wire them to the event bus and
publish them after committed state changes, or delete them.

## Findings addressed

- Medium: `azerothaccount/events.go` and `azerothstore/events.go` declare event
  types and `EventName()` methods that nothing publishes; neither plugin's
  `Config` accepts an event bus. Consumers can never observe these domain facts.

## Context

`internal/plugins/azerothaccount/events.go`,
`internal/plugins/azerothstore/events.go`, and the identity-discord plugin, which
does publish events correctly. Events must carry identifiers only, never secrets.

## Atomic change

Decide per event: publish it (if there is or will be a consumer) or delete it.
For events that are published, wire `*events.Bus` through `Config` and publish
after the transaction commits.

## Requirements

- Audit each declared event type for a real consumer or a documented near-term
  need. Delete types that have neither.
- For retained events, add `Events *events.Bus` to the plugin `Config`, wire it in
  `cmd/server/main.go`, and publish after successful commits (never before).
- Events carry IDs only (user id, account id, order id); never passwords, tokens
  or session ids.
- Ensure publishing failures cannot corrupt the request outcome (log/emit
  best-effort after the state change is durable).
- Update `docs/architecture.md`/`docs/store.md` if the event inventory changes.

## Tests

- Test: a successful account create publishes the expected event with only
  identifier fields.
- Test: a failed operation publishes nothing.
- Test: a subscriber receives the event synchronously after commit.
- Delete tests for removed event types.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green.
- No declared event type is unreferenced by a publisher.
- No event payload contains a secret (assert in tests).

## Rollback

Revert; dead types return. Prefer deletion if no consumer is planned.

## Out of scope

- Adding consumers that do new work.
- The event bus implementation.
