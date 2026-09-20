# App shell and providers

## Goal

Provide the application shell: providers, router root, query client defaults,
error boundaries and a navigation scaffold.

## Context

With the client generated, the app needs a stable composition root before
features (and the auth guard) are added.

## Requirements

- `src/app/providers.tsx`: `QueryClientProvider` (and router context), devtools
  enabled only in development.
- `src/app/query-client.ts`: sensible defaults (`staleTime`, `retry` with no
  retry on 4xx, `refetchOnWindowFocus` policy).
- `src/app/router.tsx`: TanStack Router root route with a `QueryClient` in
  context and a default not-found component.
- Root route layout: header slot, main outlet, footer slot; renders children.
- `ErrorBoundary` and a route-level error component using the typed `ApiError`.
- `src/main.tsx` mounts the router provider.

## Acceptance criteria

- App renders at `/` with the shell and a placeholder page.
- A thrown error is caught by the boundary and shown without a blank page.
- `task web:check` green; `task web:build` green.

## Implementation notes

- No domain code and no auth yet; keep the shell generic.
- Provider order documented so later tickets (auth, toasts) slot in predictably.

## Tests

- Smoke render of the shell; router context exposes the query client.

## Dependencies

- 019, 020. Unblocks 022.
