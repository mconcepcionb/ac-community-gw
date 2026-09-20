# SPA scaffold

## Goal

Stand up the SPA project under `web/` with the chosen toolchain, leaving the
existing Go workflow untouched and keeping a working test frontend during the
transition.

## Context

`web/index.html` is a self-contained test frontend served by the gateway via
`ACGW_WEB_DIR`. Vite needs `web/index.html` as its entry, so the legacy file
must move aside before scaffolding.

## Requirements

- Move the legacy frontend to `web/legacy/index.html`
  (`ACGW_WEB_DIR=./web/legacy` keeps it usable during the transition).
- Initialise the SPA at `web/`: pnpm, Vite, React, TypeScript (strict).
- `web/dist` as build output (git-ignored).
- Vite dev server proxies `/api` to `http://localhost:8080`, with
  `changeOrigin` and cookies preserved (same-origin in the browser).
- Tailwind CSS v4 + shadcn/ui init (`components.json`, `src/lib/utils.ts`).
- TanStack Router file-based routing with the Vite plugin and generated route
  tree (`src/routes/`, `routeTree.gen.ts`).
- TanStack Query provider placeholder, react-hook-form, Zod.
- Biome config (lint + format) and Vitest + React Testing Library + MSW test
  setup (`src/test/`).
- Path alias `@/*` -> `src/*` in `tsconfig` and `vite.config.ts`.
- Taskfile: `web:dev`, `web:build`, `web:test`, `web:lint`, `web:format`,
  `web:check` (biome check + `tsc --noEmit` + vitest run).
- `.gitignore`: `web/node_modules/`, `web/dist/`, `web/routeTree.gen.ts` (or
  commit it — decide and document), editor/Vite caches.

## Acceptance criteria

- `pnpm install && task web:build` produces `web/dist/index.html`.
- `task web:check` green.
- Served via `ACGW_WEB_DIR=./web/dist`, the gateway returns the built SPA at `/`.
- `task check` (Go) unchanged and green.

## Implementation notes

- No app code beyond a minimal placeholder route to prove the toolchain.
- Keep Node/pnpm versions declared (`packageManager` field, `engines`).
- The gateway does not need to know about Vite; only `web/dist` matters.

## Tests

- A trivial render test proves Vitest + RTL + jsdom are wired.

## Dependencies

- 003 (for serving `web/dist`). Unblocks 020.
