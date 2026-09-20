# Feature: admin bans and GM level

## Goal

Administer account bans and GM levels from the SPA.

## Context

Consumes `ban`, `unban` and `set_gmlevel` (ticket 012).

## Requirements

- Route `/admin/accounts` (permission-gated): lookup/table with per-account
  actions.
- Ban form (reason, duration) with confirmation; unban action; GM-level select.
- Explicit feedback and query invalidation after each action.

## Acceptance criteria

- Each action calls the matching operation; destructive actions are confirmed.
- Actions/permissions hidden appropriately.
- `task web:check` green.

## Dependencies

- 012, 020, 022, 023.
