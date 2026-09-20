# Console moderation

## Goal

Bring live moderation and account-level administration together in the console.

## Context

Today `/admin/online`, `/accounts` and `/admin/accounts` are separate. This
ticket covers use cases G1 and G2
([../../../use-cases.md](../../../use-cases.md)). Character ban/unban (G3) moved
to ticket 006, where the character context lives.

## Requirements

- Move online moderation to `/admin/online` inside the console shell: live
  player list with kick, mute and unmute.
- Merge account administration into `/admin/accounts`: list, create, set email,
  set password, ban, unban and GM level.
- Remove the old portal `/accounts` route and page.
- Gate every action by its permission and show the command result.

## Acceptance criteria

- Kick, mute, unmute, account create, set email, set password, ban, unban and GM
  level all work from `/admin/accounts` with the matching permissions.
- Actions the user lacks permission for are hidden and rejected by the API.
- The old `/accounts` path is removed and `/admin/accounts` now lives inside the
  console shell.
- `task web:check` green.

## Implementation notes

- Reuse the existing account hooks and dialogs; combine the two pages into one.
- Account actions use `azeroth.account.manage`, `azeroth.admin.accounts.ban` and
  `azeroth.admin.accounts.gmlevel`.

## Tests

- RTL + MSW for the merged management and moderation flows and their
  permission-hidden variants.

## Dependencies

- 001, 002.
