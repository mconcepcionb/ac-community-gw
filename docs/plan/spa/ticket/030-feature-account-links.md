# Feature: account links

## Goal

Manage links between community users and AzerothCore accounts.

## Context

Consumes the account-link CRUD (ticket 011).

## Requirements

- Route `/account-links` (permission-gated): table of links.
- Create link form (community user + account).
- Lookup by `user_id`, and delete with confirmation.
- Query invalidation after create/delete.

## Acceptance criteria

- CRUD flows work against mocked responses; delete asks for confirmation.
- `task web:check` green.

## Dependencies

- 011, 020, 022, 023.
