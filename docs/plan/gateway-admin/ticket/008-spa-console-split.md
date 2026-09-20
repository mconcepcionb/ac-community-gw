# SPA console route and nav split

## Goal

Mirror the API split in the SPA console: gateway-core pages under `/admin/*`,
AzerothCore pages under `/admin/azeroth/*`, with the nav grouped accordingly.

## Context

The console is flat (`web/src/routes/admin/*`):

- core: `/admin` (overview), `/admin/users`, `/admin/roles`, `/admin/audit`,
  `/admin/api-clients`, `/admin/moderation`, `/admin/store`.
- game: `/admin/accounts`, `/admin/characters`, `/admin/items`, `/admin/online`.

The SPA is not deployed, so routes move outright with no redirects.

## Requirements

- Introduce a single game descriptor, e.g. `web/src/app/games.ts`:
  `export const GAMES = [{ id: "azeroth", label: "AzerothCore" }] as const;`
- Move game route files under `web/src/routes/admin/azeroth/`:
  - `accounts/index.tsx`, `accounts/$username.tsx`
  - `characters/index.tsx`, `characters/$name.tsx`
  - `items/index.tsx`, `items/$entry.tsx`
  - `online.tsx`
- Update every `createFileRoute` path and `getRouteApi` id
  (`/admin/azeroth/accounts/`, etc.).
- Update every link to a moved page: `admin-overview-page.tsx`,
  `admin-account-detail-page.tsx`, `character-detail-page.tsx`,
  `admin-user-detail-page.tsx`, `items-page.tsx`, and any breadcrumb.
- Regroup `console-nav.tsx`: core links first, then a game-labelled section
  rendered from `GAMES`; a second game adds a section, not a rewrite.
- Regenerate `routeTree.gen.ts` (`pnpm exec vite build`).
- Update tests that navigate to the old paths.

## Acceptance criteria

- No game page is reachable under a core `/admin` path.
- The nav shows core and per-game groups; the overview cards use the new paths.
- `pnpm lint`, `pnpm typecheck`, `pnpm test` green.

## Implementation notes

- Keep the core console routes unchanged so bookmarks/links to `/admin/users`,
  `/admin/roles`, `/admin/audit` keep working.
- `getRouteApi` ids must match the new `createFileRoute` paths exactly.

## Tests

- Update `characters-page.test.tsx`, `admin-accounts-page.test.tsx`,
  `admin-account-detail-page.test.tsx`, `items-page.test.tsx`,
  `admin-online-page.test.tsx`, and `legacy-routes.test.tsx` (route table).

## Dependencies

- 002-005 (API path/operation-id changes settle first). Blocks 009.
