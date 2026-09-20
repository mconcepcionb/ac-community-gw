# Feature: admin moderation and announce

## Goal

Live server view and player moderation actions.

## Context

Consumes `GET /api/v1/azeroth/online` and the kick/mute/unmute/character-ban
endpoints (ticket 013) and `POST /api/v1/azeroth/announce` (ticket 014).

## Requirements

- Route `/admin/online`: auto-refreshing (polling) list of online players with
  per-player actions.
- Kick/mute/unmute/ban dialogs with confirmation where destructive.
- Announce form (RHF + Zod) with confirmation.
- Polling interval documented; stopped when the tab is hidden if cheap to do.

## Acceptance criteria

- Online list refreshes on an interval; actions call the right operations and
  invalidate the list.
- Announce sends and toasts; permissions gate every action.
- `task web:check` green.

## Dependencies

- 013, 014, 020, 022, 023.
