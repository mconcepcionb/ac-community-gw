# SPA frontend

## Goal

Replace the throwaway `web/index.html` test frontend with a production-grade SPA
(React + TypeScript) built on a typed, generated API contract. Prioritise
building blocks — API client, types, reusable components and the auth flow —
over final layout.

## Status

| Area | State |
| --- | --- |
| OpenAPI pipeline (`swag` -> `api/swagger.yaml`) | implemented |
| Typed backend responses for every endpoint | implemented |
| `GET /api/v1/me` with effective permissions | implemented |
| SPA fallback served by the gateway | implemented |
| SPA scaffold, shell | implemented (scaffold + typed client + app shell) |
| Auth flow (session, login, logout, guards, permissions) | implemented |
| Domain features (identity, characters, items, accounts, admin, store) | implemented |
| Production serving + legacy cleanup | implemented |

## Context

- [authentication.md](../../authentication.md) — session cookie, login flow.
- [architecture.md](../../architecture.md) — core + plugins, REST surface.
- [../../ADR/0002-net-http-only.md](../../ADR/0002-net-http-only.md) — no
  framework on the Go side, so the contract is documented by annotations.
- [../../ADR/0004-server-side-sessions.md](../../ADR/0004-server-side-sessions.md)
  — cookie sessions, no JWT for the SPA.
- [development.md](../../development.md) — Taskfile workflow.

## Scope

- A single-page app under `web/` (Vite root, build to `web/dist`).
- A committed OpenAPI spec under `api/swagger.yaml`, generated from Go
  annotations, as the source of truth for the frontend client.
- Typed request/response DTOs on every gateway endpoint, replacing ad-hoc
  `map[string]any` responses without changing the wire format.
- A generated TypeScript SDK with TanStack Query hooks and Zod validators.
- The auth flow against the existing server-side session cookie.
- Incremental feature coverage of the whole `/api/v1` surface.

## Out of scope

- Final visual design and pixel layout.
- Server-side rendering.
- Changes to the session or RBAC model beyond exposing effective permissions.
- Discord bot, guild management or anything not already in the REST API.

## Design

### Toolchain

| Concern | Choice |
| --- | --- |
| Package manager / bundler | pnpm + Vite |
| Language | TypeScript (strict) |
| Routing | TanStack Router (file-based) |
| Server state | TanStack Query |
| Auth / session | server cookie + `/api/v1/me` via TanStack Query |
| Forms | react-hook-form + Zod |
| UI | Tailwind CSS v4 + shadcn/ui (Radix) |
| Contract | swag v2 (Go) -> OpenAPI 3.1 -> `@hey-api/openapi-ts` |
| Lint / format | Biome |
| Tests | Vitest + React Testing Library + MSW |

`zustand` is intentionally not introduced; there is no client-only global state
identified yet.

### Invariants (every ticket)

- Backend: `task check` (fmt check + vet + tests + architecture boundaries) stays
  green.
- Frontend: from the scaffold onward, `task web:check` (biome + tsc + vitest)
  stays green.
- Contract: `task openapi:check` fails when `api/swagger.yaml` or the generated
  client do not match their sources.
- Converting a response from `map[string]any` to a struct MUST NOT change the
  JSON on the wire (same keys, same null/omitted semantics). Each ticket adds or
  extends a test that asserts the exact JSON.
- `operationId`s are stable (`@ID`), so adding endpoints only grows the generated
  client and never renames existing operations.

### openapi conventions

- `swag v2`, OpenAPI 3.1 output (`--v3.1`; swag v2 has no 3.0 mode), general
  API info in `cmd/server/main.go`.
- Parse `./cmd/server,./internal` with `--parseInternal`; output is
  `api/swagger.yaml`.
- One `@Tags` per plugin; `@ID` shaped as `<plugin>.<resource>.<action>`
  (for example `identity.users.list`).
- `@Success`/`@Failure` reference real DTOs; the shared error envelope is
  `httpapi.ErrorResponse`.
- Schema names are set with a trailing `@name` comment on the type
  declaration (`} // @name ErrorResponse`), not the leading doc comment, so
  generated TS type names stay clean.

### Serving strategy

> Superseded by [ADR 0012](../../ADR/0012-decoupled-spa-serving.md): the SPA is
> no longer embedded in the gateway. Ticket 035's embedding approach was
> reverted in favour of a dedicated Caddy reverse proxy.

- Development: the Vite dev server proxies `/api` to the gateway on `:8080`.
- Production: Caddy (`web/Caddyfile`, `web/Dockerfile`) serves the built SPA and
  reverse-proxies `/api/*` and the operational endpoints to the gateway, on the
  same origin.
- The legacy test frontend was removed in ticket 036.

### Dependency graph

```
001 ─┬─ 002 ─────────────┐
     ├─ 003 ────────────┤
     ├─ 004..018 ───────┤  (each unlocks its feature 024..034)
019 ─┴─ 020 ─ 021 ─ 022 ─ 023 ─ 024..034 ─ 035 ─ 036
```

A feature ticket (024..034) depends on the backend contract ticket(s) that
document its endpoints, plus 022 and 023.

## Risks

| Risk | Mitigation |
| --- | --- |
| `swag v2` is a release candidate | Use OpenAPI 3.1 (`--v3.1`, the only v3 mode); fall back to Swagger 2.0 (also consumable by hey-api) if the 3.1 generator misbehaves |
| Struct refactor changing the wire format | Exact-JSON tests per endpoint before/after |
| Generated client drift | `task openapi:check` in `task check` |
| SPA fallback breaking the JSON envelope | Keep `/api/*` handled by the mux; fallback only for non-API, extension-less paths; existing `httpapi` tests guard it |
| Migrating endpoints breaking existing consumers | Only `/me` changed, additively; the legacy test frontend was removed in 036 |
| Large surface | One ticket per resource family; features only consume already-annotated endpoints |

## Tickets

1. [001-openapi-pipeline.md](ticket/001-openapi-pipeline.md)
2. [002-auth-contract-and-me-permissions.md](ticket/002-auth-contract-and-me-permissions.md)
3. [003-spa-fallback.md](ticket/003-spa-fallback.md)
4. [004-identity-users.md](ticket/004-identity-users.md)
5. [005-azeroth-info-status.md](ticket/005-azeroth-info-status.md)
6. [006-characters-read.md](ticket/006-characters-read.md)
7. [007-characters-by-user-and-mail.md](ticket/007-characters-by-user-and-mail.md)
8. [008-items-read.md](ticket/008-items-read.md)
9. [009-accounts-write.md](ticket/009-accounts-write.md)
10. [010-accounts-read.md](ticket/010-accounts-read.md)
11. [011-account-links.md](ticket/011-account-links.md)
12. [012-admin-bans-gmlevel.md](ticket/012-admin-bans-gmlevel.md)
13. [013-admin-online-moderation.md](ticket/013-admin-online-moderation.md)
14. [014-admin-announce.md](ticket/014-admin-announce.md)
15. [015-store-catalog-read.md](ticket/015-store-catalog-read.md)
16. [016-store-catalog-write.md](ticket/016-store-catalog-write.md)
17. [017-store-wallet-orders.md](ticket/017-store-wallet-orders.md)
18. [018-store-grant.md](ticket/018-store-grant.md)
19. [019-spa-scaffold.md](ticket/019-spa-scaffold.md)
20. [020-openapi-client-codegen.md](ticket/020-openapi-client-codegen.md)
21. [021-app-shell-providers.md](ticket/021-app-shell-providers.md)
22. [022-auth-feature.md](ticket/022-auth-feature.md)
23. [023-ui-base-components.md](ticket/023-ui-base-components.md)
24. [024-feature-identity-users.md](ticket/024-feature-identity-users.md)
25. [025-feature-azeroth-status.md](ticket/025-feature-azeroth-status.md)
26. [026-feature-characters-read.md](ticket/026-feature-characters-read.md)
27. [027-feature-user-characters-mail.md](ticket/027-feature-user-characters-mail.md)
28. [028-feature-items-read.md](ticket/028-feature-items-read.md)
29. [029-feature-accounts.md](ticket/029-feature-accounts.md)
30. [030-feature-account-links.md](ticket/030-feature-account-links.md)
31. [031-feature-admin-bans.md](ticket/031-feature-admin-bans.md)
32. [032-feature-admin-moderation.md](ticket/032-feature-admin-moderation.md)
33. [033-feature-store-catalog.md](ticket/033-feature-store-catalog.md)
34. [034-feature-store-wallet-orders.md](ticket/034-feature-store-wallet-orders.md)
35. [035-production-serving.md](ticket/035-production-serving.md)
36. [036-cleanup-and-docs.md](ticket/036-cleanup-and-docs.md)
