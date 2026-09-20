# 034 — Auth flow correctness and return_to

**Phase:** 6 · **Gate:** G-standard, contributes to G6-frontend · **Depends on:** 016

## Goal

Fix the frontend auth flow: stop treating session errors as anonymous, mirror the
hardened same-origin return path, remove dead guard code, and make `/login`
idempotent under React StrictMode.

## Findings addressed

- High: `require-auth.tsx` treats any non-401 `/me` failure as anonymous, so a
  transient 5xx/network error bounces the user through the Discord OAuth loop and
  never surfaces the error.
- High: `features/auth/actions.ts` `isSafeReturnTo` mirrors the server's flawed
  helper and accepts `/\evil.com` (open redirect).
- Low: `features/auth/guard.ts` is dead code (never imported).
- Low: `/login` has no authenticated check and the effect runs twice under
  StrictMode.

## Context

`web/src/features/auth/**`, `web/src/api/client.ts`. Server-side sanitization is
fixed in ticket 016; this ticket fixes and tests the client copy.

## Atomic change

Make `RequireAuth` distinguish `anonymous` from `error`, replace the return-path
helper with a URL-based same-origin check, delete or adopt `guard.ts`, and guard
the login redirect.

## Requirements

- `RequireAuth`: redirect only when `status === "anonymous"`; for
  `status === "error"` render `ErrorState` with a `refetch` action. Optionally
  adopt `beforeLoad` guards for a flash-free experience (then remove the
  component guard or keep both consistently).
- `isSafeReturnTo`: reject `\`, parse with `new URL(value, window.location.origin)`
  and require `url.origin === window.location.origin` plus a path starting with
  `/`.
- Delete `guard.ts` if unused, or wire it and delete the duplicate component.
- `/login`: if already authenticated, redirect home; guard the effect with a
  `useRef` so StrictMode double-invocation does not double-redirect.
- Do not store tokens anywhere client-side (session stays cookie-only).

## Tests

- `RequireAuth` test: anonymous → redirect; error → error state + refetch;
  authenticated → renders children.
- `isSafeReturnTo` table test including `/\evil.com`, `//evil.com`, absolute URLs.
- `/login` test: authenticated user redirects home; effect runs once.
- Update `actions.test.ts` to cover backslashes.

## Acceptance criteria (gate G-standard, G6-frontend)

- `task web:check` green.
- A session fetch error shows an error, never a login loop.
- No backslash return path can reach an absolute URL.

## Rollback

Revert; the login loop and redirect return.

## Out of scope

- Server-side session model.
- Logout (covered by existing tests once 041 lands).
