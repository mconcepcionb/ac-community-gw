# API keys and service accounts

## Goal

Let external bots and community sites authenticate with scoped, revocable API
keys.

## Context

This ticket covers use case D1 ([../../../use-cases.md](../../../use-cases.md)).
The API is cookie-session only today.

## Requirements

- Backend: API-key authentication middleware and management endpoints (create,
  list, rotate, revoke, last-used).
- Keys are scoped to permissions and authorised by the same permission model as
  sessions.
- Console `/admin/api-clients`: manage keys; the secret is shown once.
- Key lifecycle actions are audited.

## Acceptance criteria

- A key authenticates a request and is authorised by its scopes.
- Revoking a key takes effect immediately; last-used time is visible.
- The secret is shown once and is never retrievable again.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Store only a hash of the key; prefix it for identification.
- Integrate without changing the session middleware contract.
- Scopes are validated against the permission registry, consistent with 014.

## Tests

- Go tests for authentication, scoping, revocation and audit; RTL + MSW for the
  management UI.

## Dependencies

- 001. Relates to 014 (scopes).
