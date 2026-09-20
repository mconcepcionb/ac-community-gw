# Auth feature

## Goal

Implement the authentication flow against the existing server-side cookie
session, including route protection and permission-aware helpers.

## Context

Login is a full-page redirect to `/api/v1/auth/discord/login?return_to=<path>`;
the callback sets an `HttpOnly` cookie and redirects back. The SPA learns the
session through `GET /api/v1/me`, which now returns `roles` and `permissions`
(ticket 002).

## Requirements

- `features/auth/api.ts`: `me` query helper over the generated client.
- `features/auth/use-session.ts`: `useSession()` using TanStack Query with
  `retry: false`; `401` resolves to an anonymous session (not an error), other
  errors surface.
- `features/auth/actions.ts`: `login(returnTo?)` builds the login URL with a
  safe same-site `return_to`; `logout()` calls `POST /api/v1/auth/logout`,
  clears the `me` query and redirects.
- `features/auth/require-auth.tsx`: `<RequireAuth>` component and a
  `beforeLoad` guard that redirects anonymous users to login with
  `return_to` set to the current location.
- `features/auth/use-permissions.ts`: `usePermissions()` and
  `<Can permission="…" fallback>` reading the effective permissions.
- Session states: `loading`, `anonymous`, `authenticated`.
- Integrate with the shell: show login/logout and the principal summary.

## Acceptance criteria

- Anonymous users hitting a protected route are sent through the login redirect.
- After a mock `me` response, protected content renders; permissions gate
  actions.
- Logout clears state and returns to the anonymous shell.
- `task web:check` green.

## Implementation notes

- Never store tokens in JS; rely on the `HttpOnly` cookie (credentials included
  in the shared client).
- `return_to` must be validated client-side to a same-site absolute path before
  being sent.
- Do not cache `me` across logout; explicitly invalidate.

## Tests

- MSW: anonymous (`401`), authenticated with permissions, `403`-driven UI
  hiding, logout invalidation, and the login URL/`return_to` construction.

## Dependencies

- 002, 020, 021. Unblocks every feature ticket.
