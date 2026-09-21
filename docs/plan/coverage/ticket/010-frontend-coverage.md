# Frontend coverage

**Milestone:** D · **Gate:** G-standard · **Depends on:** 001

## Goal

Add Vitest coverage plumbing with a 60% statements threshold and cover the
highest-value untested frontend areas.

## Context

`@vitest/coverage-v8` is absent from `web/package.json` and the lockfile; there
is no `coverage` block in `web/vite.config.ts` and no `test:coverage` script.
33 test files exist. Entirely untested features include `accounts`,
`account-links`, `leaderboards`, `home` and `audit`; `web/src/routes` has no
direct tests. Generated code to exclude: `web/src/api/generated/**` and
`web/src/routeTree.gen.ts` (already excluded by `web/biome.json:8-9`).

## Requirements

- Add `@vitest/coverage-v8`; add `test:coverage` (`vitest run --coverage`).
- Add a `coverage` block to `web/vite.config.ts`: provider `v8`, `include:
  ["src/**"]`, exclude `src/api/generated/**`, `src/routeTree.gen.ts`,
  `src/test/**`, `src/**/*.d.ts`; thresholds 60 statements (and lines), with
  functions/branches added when reached.
- Add `task web:coverage:check` and a CI step in the web job.
- Fill high-value gaps only (behavioural gaps stay in
  `review-remediation/041`):
  - security-sensitive `accounts` dialogs (create, set-email, set-password) and
    `account-links` dialogs;
  - `leaderboards` and `home` pages;
  - `audit/entity-history`;
  - `api/errors.ts`, `hooks/use-debounced-value`, `lib/utils`.

## Tests

- The new component/hook tests.
- A check that the coverage run fails when the threshold is not met.

## Acceptance criteria

- `task web:check` and `task web:coverage:check` green at ≥ 60% statements on
  the included set.
- Generated client and `routeTree.gen.ts` are excluded from the report.

## Out of scope

- The full behavioural gap list in `review-remediation/041`.
- Visual/E2E testing.
