# Player reports

## Goal

Let an authenticated player report another player so staff can act on it.

## Context

There is no reporting capability today. This is a small new feature that feeds
the moderation queue (015) and reflects the "player reports" item in use case G5
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- Backend: a report submission endpoint (reporter, target character or account,
  category, free-text message) with validation, storage and a status
  (`open`/`closed`).
- Rate-limited per user to deter spam; the reporter is recorded; audited.
- Portal: a report form (for example `/report`) reachable from the portal,
  requiring a signed-in account.
- Reports are readable by staff through the queue (015); the reporter sees only
  their own submissions.

## Acceptance criteria

- A signed-in player can submit a report against a target and see it listed as
  open.
- A report cannot be submitted anonymously; spam is rate-limited.
- Staff can read and close reports via the queue.
- Endpoints are annotated and the client regenerated.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Keep the model minimal: no threaded conversation in v1.
- The target is stored as identifiers, never free text alone.

## Tests

- Go tests for submission, validation, rate limit and staff reads.
- RTL + MSW for the portal form and the "my reports" list.

## Dependencies

- 001, 002, 008. Consumed by 015.
