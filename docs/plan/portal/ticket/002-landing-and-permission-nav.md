# Landing and permission navigation

## Goal

Route users to the right surface after sign-in, build navigation from
permissions, and make console routes unreachable for players.

## Context

Today `app-nav.tsx` lists every feature behind `<PermissionGate>` with no notion
of surfaces. The `/me` payload already returns effective permissions.

## Requirements

- Landing resolver: after sign-in, a user holding any console permission lands
  on `/admin`; everyone else lands on `/`.
- Surface-specific navigation built from `/me` permissions.
- A portal-only user hitting `/admin/*` resolves to a forbidden state with no
  console chrome, not a redirect loop.
- A user with both surfaces can move between portal and console from a
  persistent link.
- Replace `app-nav.tsx` with the surface-aware nav.

## Acceptance criteria

- Staff, player and anonymous users land per the resolver.
- Links the user cannot use are not rendered; direct navigation is forbidden.
- A console user can move between surfaces.
- `task web:check` green.

## Implementation notes

- Add a `CONSOLE_PERMISSIONS` list and a `hasAnyConsolePermission(p)` helper;
  keep it in one place so 003-007 reuse it.
- Authorization stays server-side; this is navigation only.

## Tests

- Landing resolver (staff, player, anonymous) and the forbidden state.

## Dependencies

- 001.
