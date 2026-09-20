# Structured online list and moderation autocomplete

## Goal

Show the online players as structured, actionable rows and replace the free-text
moderation target with an autocomplete.

## Context

- The page renders the raw console output in a `<pre>`
  (`web/src/features/admin/admin-online-page.tsx:29-33`).
- The backend runs `.account onlinelist`
  (`internal/plugins/azerothadmin/moderation.go:53-56`), which AzerothCore
  gates behind GM security level 4 (SEC_CONSOLE). With a lower SOAP account it
  replies "Command '.account onlinelist' does not exist".
- The moderation panel takes free-text player/character names and always renders
  Kick/Mute/Unmute/Ban/Unban (`web/src/features/admin/moderation-panel.tsx:233-262`).

## Requirements

- Backend: provide a structured online-player read. Preferred: query the
  character DB (`SELECT guid, name, level, race, class, ... FROM characters
  WHERE online = 1`) exposed at the existing `GET /api/v1/azeroth/online`,
  changing the response from `{ output }` to a typed list. Fallback: keep the
  SOAP command but require/detect GM level 4 and surface an actionable error.
- Return per-player: name, level, class, race, guild, account username, online
  since if available.
- Console: render a sortable `DataTable` with per-row Kick/Mute/Ban actions;
  keep a manual **Refresh** and show the last refresh time.
- Moderation target: replace the free-text input with the shared `Autocomplete`
  backed by the character search (007); only enable actions for a selected
  character; state-gate Ban xor Unban.

## Acceptance criteria

- The online page shows structured rows, not a text blob.
- The list works without requiring the SOAP account to hold console level 4
  (if the DB path is chosen) and fails with a clear message otherwise.
- The moderation input offers suggestions and requires a selected value.
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- The structured read also enables per-player actions and future "last seen"
  data; prefer it over parsing console text.
- This replaces the `{ output }` response with a typed list; pre-production, so
  the shape changes outright. Regenerate the OpenAPI client, update
  `use-online.ts` and the generated types in the same change.

## Tests

- Go tests for the new response shape and empty/large lists.
- Update `admin-online-page.test.tsx`; add autocomplete interaction tests.

## Dependencies

- 001, 007 (autocomplete source). Relates to the existing moderation panel.
