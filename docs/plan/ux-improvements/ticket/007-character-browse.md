# Global character browse

## Goal

Let staff browse all characters with pagination and filters by account and
character name, instead of being forced to type an account first.

## Context

- `handleListCharacters` requires `account` or `account_id` and returns
  `422 missing_account` otherwise
  (`internal/plugins/azerothcharacter/characters.go:86-90, 174-199`).
- The SPA gates the query on an account and hides the table entirely until one
  is entered (`web/src/features/characters/use-characters.ts:18`,
  `characters-page.tsx:92-133`), making the name filter useless on arrival.
- No response includes a total count (`CharactersResponse` is name-only,
  `types.gen.ts:178-180`).

## Requirements

- Backend: make `account`/`account_id` optional. When absent, list all
  characters (respecting existing filters and a conservative default limit).
  Guard the global path with `azeroth.character.list`.
- Add a `total` field to the list response using a `COUNT(*)` query with the
  same predicates.
- Filters: `account` (username, exact/resolved), `filter` (name substring),
  `limit`, `offset`. Consider an `online` filter.
- Console page: always render the table; account filter becomes the shared
  autocomplete (001); name filter stays debounced; rows link to the detail
  page; header shows the total.
- Keep the "characters of an account" deep link working (storefront/user-360).

## Acceptance criteria

- `/admin/characters` lists characters with no account typed.
- Typing an account narrows to that account; typing a name filters within it.
- The range label reflects the real total, not "page full / not full".
- `task check`, `task openapi:check`, `task web:check` green.

## Implementation notes

- The `COUNT(*)` runs against the same world DB; ensure the name index is used
  (`migrations/00008` covers gateway tables only; verify AzerothCore schema).
- Reuse the SPA pagination helper from 001 and align with 006.

## Tests

- Go tests: no-account listing, account listing, count correctness.
- Update `characters-page.test.tsx` for the always-visible table.

## Dependencies

- 001 (autocomplete, pagination). Feeds 008 and 010.
