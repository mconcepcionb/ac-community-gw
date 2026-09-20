# 037 — Route parameter and query error states

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 001

## Goal

Stop infinite loading on invalid route parameters and stop showing wrong data
when a query fails.

## Findings addressed

- Medium: `use-items.ts`/`item-detail-page.tsx` disable the query for an invalid
  or zero entry, but React Query reports a disabled query as `isPending`, so
  `/items/abc` or `/items/0` renders `LoadingState` forever. The same pattern
  exists in `use-account-links.ts`, `use-products.ts`, `use-user-characters.ts`.
- Medium: `wallet-page.tsx` handles only `isPending`; `wallet.isError` falls
  through to `wallet.data?.balance ?? 0`, showing a false `0` balance.
- Nit: `items-page.tsx` treats `class=0` as falsy, so the input shows empty while
  the filter is active.
- Nit: pagination "Next" stays enabled when the page is exactly full, leading to
  an empty next page.

## Context

`web/src/features/items/**`, `account-links`, `store`, `characters`,
`use-account-links.ts`, `use-products.ts`, `use-user-characters.ts`.

## Atomic change

Distinguish "invalid parameter" from "loading" and render explicit error/empty
states; fix the falsy-zero and pagination edge cases.

## Requirements

- For each route-param hook: detect a non-finite/≤0/invalid id before the query
  and render `NotFound`/`EmptyState` (not loading); keep the query disabled.
- `wallet-page`: render `ErrorState` when `isError`; never substitute `0`.
- Address inputs: treat `0` as a valid value (`search.class !== undefined` /
  `?? ""` correctly).
- Pagination: enable "Next" based on whether the current page returned a full
  `limit`, and disable "Previous" on the first page.
- Apply consistently across items/accounts/characters/users/account-links pages.

## Tests

- Test: `/items/abc` and `/items/0` render a not-found state, not loading.
- Test: a wallet query error renders an error, not `0`.
- Test: `class=0` displays `0` in the input.
- Test: full last page disables "Next".

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green.
- No route with an invalid parameter renders an indefinite spinner.

## Rollback

Revert individual page fixes; low risk.

## Out of scope

- Backend pagination semantics (ticket 018).
