# Feature: user characters and mail

## Goal

Show the characters of a community user and send them in-game mail.

## Context

Consumes `GET /api/v1/azeroth/users/{user_id}/characters` and
`POST /api/v1/azeroth/mail` (ticket 007).

## Requirements

- User characters section reachable from the identity user detail/list.
- Mail form (RHF + Zod): character, subject, body, items, money; validation
  mirrors the backend DTO.
- On success: toast + invalidation of affected queries; on failure: field and
  form errors from `ApiError.details`.
- Destructive/irreversible warning before sending.

## Acceptance criteria

- Submitting a valid form calls `azeroth.mail.send` and shows a success toast.
- Backend validation errors are surfaced without losing form state.
- `task web:check` green.

## Dependencies

- 007, 020, 022, 023.
