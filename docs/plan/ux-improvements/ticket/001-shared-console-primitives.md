# Shared console primitives

## Goal

Extract the repeated console interaction patterns into a small set of shared
components so every later ticket is a smaller diff and the console behaves
consistently.

## Context

The review found the same problems copied across pages:

- Both positive and negative actions rendered unconditionally: accounts
  (`web/src/features/admin/admin-accounts-page.tsx:64-81`), moderation panel
  (`web/src/features/admin/moderation-panel.tsx:251-258`), character detail
  (`web/src/features/admin/character-ban-actions.tsx:76-119`).
- Status rendered as `yes`/`no` text: `online` and `banned` in accounts,
  characters and the user-360.
- Duplicated offset-pagination markup with no totals: accounts
  (`admin-accounts-page.tsx:150`), characters (`characters-page.tsx:102`),
  users (`users-page.tsx:80`), items (`items-page.tsx:107`).
- Free-text identifiers where a lookup is possible: moderation name
  (`moderation-panel.tsx:240`), store grant (`grant-dialog.tsx:78-79`),
  create-link, mail character/item ids (`mail-form.tsx`).

## Requirements

- `StatusBadge`: renders online/offline, banned/active, claimed/unclaimed with
  consistent colours and accessible text (not colour-only).
- `RowActions`: lays out per-row actions and supports state-gated actions
  (render `Ban` xor `Unban`).
- `Pagination`: a controlled strip showing a page-size selector, `Prev`/`Next`
  and a range/total label; URL-search-backed where the route already uses
  search params.
- `usePagination` hook: derives `limit`/`offset` from route search and exposes
  navigation helpers.
- `Autocomplete`: debounced async input with keyboard selection, loading/empty
  states, accessible listbox semantics, and a pluggable option loader.
- Migrate the five existing pages to the new primitives without changing their
  query contracts (list tickets keep their own pagination work).

## Acceptance criteria

- The duplicated pagination blocks are removed; pages use `Pagination`.
- Every status cell uses `StatusBadge`.
- No row renders an action that does not apply to its state.
- `Autocomplete` is covered by RTL tests (typing, selection, keyboard, empty).
- `task web:check` and `task web:test` green.

## Implementation notes

- Keep components presentational; queries stay in feature hooks.
- `Autocomplete` must work with the generated TS client options (query keys and
  `enabled`), not bespoke fetch calls.
- Preserve existing `aria-label`s so current tests keep passing where the
  markup moved.

## Tests

- RTL + MSW unit tests for `StatusBadge`, `Pagination` and `Autocomplete`.
- Update `data-table.test.tsx` if the shared table API changes.

## Dependencies

- Defensive ordering: lands first; unblocks 005-011.
