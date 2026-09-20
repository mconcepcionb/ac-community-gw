# Account list moderation controls

## Goal

Make the accounts list read at a glance and offer exactly the action that
applies to each row.

## Context

`web/src/features/admin/admin-accounts-page.tsx` currently renders:

- a plain `gm_level` number in one column **and** a separate always-present
  "GM level" button (`:26`, `:82-91`);
- `banned` as `yes/no` **and both** a `Ban` and an `Unban` button
  (`:32-36`, `:64-81`);
- `email`, `online`, `last_login` but not `ban_reason`, `last_ip` or
  `expansion`, which the API already returns (`web/src/api/generated/types.gen.ts:114-125`).

## Requirements

- **GM column is the control.** Show the current level as a button (e.g.
  `3 — Administrator`); opening it pre-fills `SetGmLevelDialog` with the current
  level and realm.
- **Banned column is the control.** Render a `StatusBadge`; show `Unban` (with
  confirmation) when banned and `Ban` (with the duration/reason dialog) when
  not. Never both.
- Show `ban_reason` (tooltip or subtext) and `last_ip`; keep `last_login` and
  `online` badges.
- Link `username` to the account detail page (005); link `claimed_by` to the
  user detail page.
- Use the shared `StatusBadge`, `RowActions` and `Pagination` from 001.
- Keep the debounced filter and URL-backed pagination; add totals once 007's
  count pattern exists (or reuse the same helper).

## Acceptance criteria

- A banned account shows only `Unban`; an unbanned account shows only `Ban`.
- The GM button reflects the account's current level when opened.
- Claim status is visible and links to the owner.
- Existing `admin-accounts-page.test.tsx` is updated and green; new tests cover
  the state-gated actions.

## Implementation notes

- `SetGmLevelDialog` currently defaults to level `0` and realm `""`
  (`web/src/features/admin/set-gmlevel-dialog.tsx:57`); add `currentLevel` /
  `currentRealm` props.
- `UnbanAccountButton` is already a confirmation component
  (`web/src/features/admin/unban-account-button.tsx`); reuse it and gate it.

## Tests

- RTL + MSW for: banned→Unban only, unbanned→Ban only, GM prefill, claim links.

## Dependencies

- 001, 005 (detail link/claim column).
