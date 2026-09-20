# Frontend

The SPA lives under `web/` and talks to the gateway over `/api/v1`. It is a
React + TypeScript application built with Vite.

## Toolchain

| Concern | Choice |
| --- | --- |
| Package manager / bundler | pnpm + Vite |
| Language | TypeScript (strict) |
| Routing | TanStack Router (file-based) |
| Server state | TanStack Query |
| Forms | react-hook-form + Zod |
| UI | Tailwind CSS v4 + shadcn/ui (Radix) |
| Contract client | `@hey-api/openapi-ts` (SDK + React Query + Zod) |
| Lint / format | Biome |
| Tests | Vitest + React Testing Library + MSW |

## Structure

```
web/
  openapi-ts.config.ts   hey-api codegen config
  src/
    api/
      client.ts          configured fetch client + ApiError interceptor
      errors.ts          ApiError, isApiError, isUnauthorized
      generated/         committed output of `task openapi:client`
    app/                 providers, router, shell, nav
    components/
      ui/                shadcn primitives
      common/            DataTable, states, dialogs, form controls
    features/            one folder per domain (auth, identity, characters, …)
    hooks/               small shared hooks
    routes/              TanStack Router file-based routes
    test/                MSW server and setup
```

## Contract and code generation

The Go handlers are annotated with `swag` comments. `task openapi` regenerates
`api/swagger.yaml` (OpenAPI 3.1) and the typed client under
`web/src/api/generated/`. Both are committed; `task openapi:check` fails when
they drift.

- `operationId` is stable (`@ID` in the Go annotations), so generated names do
  not change when endpoints are added.
- Schema names are set with a trailing `@name` comment on the Go type.
- Never hand-edit `web/src/api/generated/`.

Feature code imports from `@/api` (barrel that re-exports the generated SDK,
React Query options and the normalised `ApiError`).

## Authentication

The gateway owns the session with an `HttpOnly` cookie; the SPA never stores
tokens.

- `useSession()` queries `GET /api/v1/me`; a `401` resolves to an anonymous
  session (not an error).
- `login(returnTo)` redirects to `/api/v1/auth/discord/login` with a validated
  same-site `return_to`.
- `logout(queryClient)` revokes the session and invalidates the principal.
- Protected routes render `<RequireAuth>`; route loaders can use
  `requireSessionFor(queryClient)`.
- Permission-aware UI uses `usePermissions()` / `<Can permission>` /
  `<PermissionGate permission>` against the effective permissions returned by
  `/me`.

## Serving

The SPA is a separate artifact; the gateway never serves it (ADR 0012).

- Development: `task dev` runs the gateway and Vite together. Vite serves the
  SPA on `:5173` with HMR and proxies `/api` to the gateway on `:8080`.
- Production: Caddy (`web/Caddyfile`, `web/Dockerfile`) serves the built SPA
  and reverse-proxies `/api/*`, `/healthz`, `/readyz` and `/metrics` to the
  gateway. Keeping both on one origin means the session cookie and the Discord
  OAuth callback work unchanged, with no CORS.
- `task docker:up` runs the whole stack (Caddy + gateway + PostgreSQL);
  `task web:image` builds the SPA/proxy image on its own.

## Testing

```bash
task web:check   # biome + tsc + vitest
```

- MSW intercepts the generated SDK; the default `/api/v1/me` handler returns an
  anonymous session unless a test overrides it.
- Route-level tests mount the real router with an in-memory history.
- `src/test/setup.ts` provides jsdom shims (AbortSignal, ResizeObserver,
  scrollTo) required by the generated client and Radix components.
