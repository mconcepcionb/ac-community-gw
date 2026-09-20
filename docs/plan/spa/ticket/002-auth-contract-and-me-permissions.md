# Auth contract and `/me` permissions

## Goal

Document the authentication endpoints and expose the principal's effective
permissions so the SPA can render permission-aware UI.

## Context

`GET /api/v1/me` currently returns `user_id`, `discord_id` and `roles`. The RBAC
grants live in the `Authorizer`; `Authorizer.Permissions(roles)` already computes
the union but does not actually sort despite its doc comment.

## Requirements

- `Authorizer.Permissions` returns a deterministically sorted slice.
- Inject a permission lister into `identity-discord` (small interface accepting
  `[]permissions.Role` and returning `[]permissions.Permission`; wired in
  `cmd/server`).
- `handleMe` returns typed `MeResponse{ user_id, discord_id, roles, permissions }`;
  `roles` and `permissions` are always arrays (never `null`).
- Typed DTOs for `GET /auth/discord/login`, `GET /auth/discord/callback`,
  `POST /auth/logout` and `GET /me`.
- swag annotations for all four endpoints with stable `@ID`s under the `auth`
  tag, including `302` responses for login/callback and the error envelope.

## Acceptance criteria

- `/me` is unchanged for existing consumers except for the additive
  `permissions` field.
- Anonymous `/me` still returns `401 unauthorized`.
- `logout` still returns `204`.
- Annotated operations appear in `api/swagger.yaml`.

## Implementation notes

- Follow the existing plugin rule: `identity-discord` may import core
  `permissions` (a core contract); no adapter import.
- Sorting is done in `Authorizer.Permissions`, not in the handler, so every
  caller benefits. Add a deterministic test there too.

## Tests

- `permissions`: `Permissions` returns a sorted, de-duplicated union.
- `identitydiscord`: `/me` JSON contains `roles` and `permissions` arrays for a
  principal with grants, and empty arrays when there are none.
- Existing oauth handler tests keep passing.

## Dependencies

- 001 (annotation pipeline). Unblocks 022 and 024+.
