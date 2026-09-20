# Console moderation

## Goal

Bring live moderation and account-level administration together in the console
as one area.

## Context

Today `/admin/online`, `/accounts` and `/admin/accounts` are separate pages.
This ticket covers use cases G1, G2 and G3
([../../../use-cases.md](../../../use-cases.md)).

## Requirements

- Move online moderation to `/admin/online` inside the console shell: live
  player list with kick, mute and unmute.
- Move account ban/unban and GM-level changes to `/admin/accounts`.
- Surface character ban/unban (G3): the endpoint exists but has no UI today.
- Gate every action by its permission and show the command result.

## Acceptance criteria

- Kick, mute, unmute, account ban/unban, GM level and character ban all work from
  the console with the matching permissions.
- Actions the user lacks permission for are hidden and rejected by the API.
- The old `/accounts` path is removed and `/admin/accounts` now lives inside the
  console shell.
- `task web:check` green.

## Implementation notes

- Reuse the existing hooks and dialogs; only rehome them and add the
  character-ban control.
- Character ban uses `azeroth.admin.characters.ban`; account actions use
  `azeroth.admin.accounts.ban` / `.gmlevel`.
- Every action is audited by the backend; no frontend audit work.

## Tests

- RTL + MSW for each action and its permission-hidden variant.

## Dependencies

- 001.
