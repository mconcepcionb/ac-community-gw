# SPA portal route split

## Goal

Mirror the API split in the player portal: gateway-core pages at the root,
AzerothCore pages under `/azeroth/*`.

## Context

The portal (`web/src/routes/_portal/*`, a pathless layout) is flat. Core pages
are `/`, `/login`, `/profile`, `/wallet`, `/store`, `/report`, `/forbidden`.
Game pages are `/characters` (my characters), `/leaderboards`, `/status`,
`/onboarding` (create/claim a game account).

## Requirements

- Move game pages under `web/src/routes/_portal/azeroth/`:
  - `characters/index.tsx`
  - `leaderboards/index.tsx`, `leaderboards/$board.tsx`
  - `status.tsx`
  - `onboarding.tsx`
- Update every `createFileRoute` path and `getRouteApi` id.
- Update portal navigation and links: `app-nav.tsx`, `home-page.tsx`,
  `portal-layout.tsx`, the onboarding card, the profile page, and any
  `Link to="/characters"`, `/status`, `/leaderboards`, `/onboarding`.
- Group the portal nav/home game entries under a game-labelled section using
  `GAMES` from ticket 008.
- Regenerate `routeTree.gen.ts`.
- Update tests that navigate to the old paths.

## Acceptance criteria

- No game page is reachable at a core portal path.
- The portal nav/home group core and per-game entries.
- `pnpm lint`, `pnpm typecheck`, `pnpm test` green.

## Implementation notes

- `/store` and `/wallet` stay core (the store is a gateway feature; see ticket
  007).
- `/report` stays core: reports are a gateway feature
  (`gw.report.*`) even though the target is a game character.
- Keep `/` as the core home; game widgets move into a labelled section.

## Tests

- Update `app-shell.test.tsx`, `onboarding-page.test.tsx`,
  `my-characters-page.test.tsx`, `status-page.test.tsx`, and the leaderboards
  tests.

## Dependencies

- 008 (shares the `GAMES` constant and nav pattern).
