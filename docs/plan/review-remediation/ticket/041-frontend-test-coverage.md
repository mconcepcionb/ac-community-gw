# 041 — Frontend test coverage and harness

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 035, 036, 037, 038

## Goal

Close the frontend test gaps and make the MSW harness safe for multiple tests.

## Findings addressed

- Low: no tests for `RequireAuth`, logout, the product update path, ban/unban/
  GM-level, delete-product/link, grant, mute/unmute, account-link detail,
  pagination, or query invalidation after mutations.
- Low: `harness.test.tsx` is a placeholder.
- Low: module-level capture vars (`banBody`, `createBody`, `purchaseBody`) are not
  reset between tests, which will break when a second test is added.
- Nit: `openapi-ts.config.ts` enables the Zod plugin but `zod.gen.ts` is never
  imported and no `responseValidator` is set.

## Context

`web/src/test/**`, feature tests, `openapi-ts.config.ts`.

## Atomic change

Add the missing tests, reset shared state, and either use the generated Zod
validators or disable the plugin.

## Requirements

- Reset all module-level capture variables in `beforeEach` (or move them into the
  test scope).
- Add tests for the untested flows listed above, focusing on behaviour and exact
  request payloads.
- Either wire `responseValidator` from `zod.gen.ts` into the query client (giving
  runtime response validation) or remove the `zod` plugin from
  `openapi-ts.config.ts`; do not leave it unused.
- Replace the `harness.test.tsx` placeholder with a real smoke test or remove it.
- Keep MSW `onUnhandledRequest: "error"`.

## Tests

- The new tests fail on the pre-fix code (for the behaviour tickets) and pass
  afterward.
- Running the suite twice in a row yields the same result (no leaked state).

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green.
- The listed flows have coverage.
- No unused codegen plugin.

## Rollback

Revert tests; no production impact.

## Out of scope

- Visual regression testing.
