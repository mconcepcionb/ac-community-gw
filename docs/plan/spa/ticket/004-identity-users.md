# Identity: list users

## Goal

Give `GET /api/v1/identity/users` a typed response and document it.

## Context

`handleListUsers` in `internal/plugins/identitydiscord/users.go` returns an
ad-hoc `map[string]any` built by `userJSON`.

## Requirements

- Introduce `ListUsersResponse{ users []User }` and `User` structs with JSON
  tags identical to the current keys: `user_id`, `discord_id`, `username`,
  `global_name`, `display_name`, `created_at`.
- `users` is always an array, never `null`.
- swag annotations: `@Tags identity`, `@ID identity.users.list`,
  query params `filter`, `limit`, `offset`; `@Success 200`,
  `@Failure 503` and the shared error envelope.
- Annotate the required permission (`identity.users.list`) in the description.

## Acceptance criteria

- Response JSON is byte-for-byte equivalent to the previous output for the same
  input (asserted on a fixed user).
- Operation `identity.users.list` appears in `api/swagger.yaml`.
- `task check` green.

## Implementation notes

- Reuse `userdir.User`; the DTO only fixes the JSON representation.
- `created_at` stays RFC3339 in UTC.

## Tests

- Extend `http_test.go` to assert the exact JSON envelope for a fake repository.

## Dependencies

- 001.
