# API keys and service accounts

## Goal

Let external bots and community sites authenticate with scoped, revocable API
keys, chosen through a generated permission picker.

## Context

This ticket covers use case D1 ([../../../use-cases.md](../../../use-cases.md)).
The API is cookie-session only today. There is no endpoint that lists the
registered permissions, which the scopes form needs.

## Requirements

- Backend: a **list-registered-permissions** endpoint exposing each permission's
  name, description and owner (from the permission registry).
- Backend: API-key authentication middleware and management endpoints (create,
  list, rotate, revoke, last-used).
- A key's scopes are a subset of the registered permissions, authorised through
  the same permission model as sessions.
- Console `/admin/api-clients`: a generation form whose scope picker is
  **generated from the permissions endpoint** (grouped by owner), plus list,
  rotate and revoke. The secret is shown once.
- Key lifecycle actions are audited.

## Acceptance criteria

- The generation form lists every registered permission and lets the operator
  select a subset.
- A key authenticates a request and is authorised by its scopes.
- Revoking a key takes effect immediately; last-used time is visible.
- The secret is shown once and is never retrievable again.
- `task check`, `task openapi:check` and `task web:check` green.

## Implementation notes

- Store only a hash of the key; prefix it for identification.
- Integrate without changing the session middleware contract.
- Scopes validate against the permission registry, consistent with 017.

## Tests

- Go tests for authentication, scoping, revocation and audit; RTL + MSW for the
  management UI and the generated scope picker.

## Dependencies

- 002, 017.
