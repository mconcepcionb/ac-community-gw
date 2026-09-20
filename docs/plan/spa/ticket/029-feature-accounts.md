# Feature: accounts

## Goal

List AzerothCore accounts and perform account write operations.

## Context

Consumes `GET /api/v1/azeroth/accounts` (ticket 010), and the create, password
and email operations (ticket 009).

## Requirements

- Route `/accounts` (permission-gated): searchable/paginated table.
- Create-account dialog/form (RHF + Zod).
- Per-account actions: set password, set email, with confirmation where
  destructive.
- Invalidate the account list after each mutation; surface `ApiError`.
- Unavailable state when the login DB is not configured (503).

## Acceptance criteria

- List renders from mocked data; each mutation calls the right operation and
  refreshes the list.
- Actions are hidden without the matching permission.
- `task web:check` green.

## Dependencies

- 009, 010, 020, 022, 023.
