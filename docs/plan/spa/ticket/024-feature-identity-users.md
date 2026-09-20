# Feature: identity users

## Goal

List and search community users in the SPA.

## Context

Consumes `GET /api/v1/identity/users` (ticket 004) and the generated
`identity.users.list` operation.

## Requirements

- Route `/identity/users` (protected, permission `identity.users.list`).
- Search input (filter) with debounce, pagination (`limit`/`offset`).
- `DataTable` columns: user_id, discord_id, username, global_name,
  display_name, created_at.
- Loading, empty and error states via the common components.

## Acceptance criteria

- List renders from a mocked response; searching and paging update the query
  params and cache keys.
- Users without the permission cannot reach the route (guard) and see no
  entry point in navigation.
- `task web:check` green.

## Dependencies

- 004, 020, 022, 023.
