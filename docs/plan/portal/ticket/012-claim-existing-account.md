# Claim an existing account with an in-game code

## Goal

Let a signed-in user prove ownership of an existing game account and link it.

## Context

This ticket covers use case P3 ([../../../use-cases.md](../../../use-cases.md))
and extends the onboarding entry from 011. It is a **Now** horizon ticket.

## Requirements

- Backend: a start-claim endpoint that sends a one-time code in-game to the
  account holder, and a submit-code endpoint that validates it within a window.
- Backend: staff visibility of pending and failed claims.
- Requires Discord guild membership; rate-limited per user and per IP; audited.
- Reject an account already linked to another user.
- Portal `/onboarding`: the claim path with code entry, retry and expiry
  feedback.

## Acceptance criteria

- A correct code links the account and reveals its characters.
- A wrong or expired code fails safely and can be retried; rate limits apply.
- A claim on an account linked to another user is rejected.
- Staff can see pending and failed claims.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Delivery uses the in-game mail/command capability owned by `azeroth-character`
  or `azeroth-admin`; the claim owner plugin manages the code lifecycle.
- Store only a hash of the code with its expiry.
- One account per community user, consistent with 011.

## Tests

- Go tests: issue, correct, wrong, expired, already linked, rate limit.
- RTL + MSW for the claim flow.

## Dependencies

- 011. 015 surfaces pending claims in the queue.
