# 038 — API error and URL-state robustness

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 001

## Goal

Make error rendering and URL-driven filter state robust, and document/align the
retry policy.

## Findings addressed

- Low: `errors.ts` stringifies non-envelope JSON objects to `"[object Object]"`.
- Low: filter inputs are initialized once from `search`, so browser back/forward
  changes the URL but not the input (accounts, characters, items, users,
  admin-accounts).
- Low: `useSession`/`useStatus` use `retry:false` while other queries retry 5xx,
  without a documented rationale.
- Nit: `announce-form.tsx` has a dead `onSubmit={(e) => e.preventDefault()}` and no
  submit button, so keyboard submit is impossible.
- Low/Medium security: with `ACGW_SESSION_COOKIE_SAMESITE=none`, mutating
  endpoints become CSRF-able (no CSRF token).

## Context

`web/src/api/errors.ts`, `query-client.ts`, `hooks/use-debounced-value.ts`,
feature filter pages, `web/src/api/client.ts`.

## Atomic change

Harden error parsing, sync inputs with router search state, align retry policy,
fix the announce form, and add CSRF defense for `SameSite=None`.

## Requirements

- `errors.ts`: fall back to `body.message`/`body.code`/`JSON.stringify` (truncated)
  instead of `[object Object]`.
- Filter inputs: drive from router search state (`useEffect`/controlled) so
  back/forward re-syncs; preserve debounce.
- `query-client.ts`: document the retry policy centrally; make session/status
  opt-outs explicit and commented (or align them).
- `announce-form`: remove the dead handler or add a working submit path
  (keyboard-accessible).
- CSRF: when `SameSite=None` is configured, require a double-submit token on
  mutating requests; otherwise document that `SameSite=Lax` is mandatory for this
  deployment.
- Do not introduce tokens in `localStorage`.

## Tests

- `errors.ts` test: a non-envelope object yields a readable message.
- Test: changing the URL back/forward updates the filter input.
- Test: announce is submittable via keyboard.
- CSRF test: a mutating request includes the token when configured.

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green.
- No user-visible `[object Object]`.
- Filter inputs reflect the current URL.

## Rollback

Revert individual fixes.

## Out of scope

- Reworking list layouts.
- The confirm-dialog fix (ticket 036).
