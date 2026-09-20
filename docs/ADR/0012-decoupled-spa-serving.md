# ADR 0012: Decouple the SPA from the gateway binary

## Status

Accepted

Supersedes the *serving* decision of [ADR 0011](0011-spa-contract-and-serving.md).
The contract and session decisions of ADR 0011 remain in force.

## Context

ADR 0011 embedded the built SPA in the gateway binary with `//go:embed` under
the `embedui` build tag, so production shipped a single self-contained artifact.
In practice this couples two things that have different lifecycles:

- The gateway is a multi-usage REST API. Other consumers (Discord bots, other
  frontends, scripts) use it without this SPA.
- Embedding makes the API image and CI depend on the SPA build, forces every
  frontend release to ride along with an API release, and prevents hosting,
  caching or versioning the SPA independently.

The SPA is already decoupled in development: `task dev` runs the gateway and
Vite side by side, with Vite proxying `/api` to the gateway. Only production
was coupled.

## Decision

- **The gateway does not serve the SPA.** `internal/core/webui`, the `embedui`
  build tag, the `WebFS` dependency, `ACGW_WEB_DIR` and the `/` SPA handler are
  removed. The gateway serves `/api/*` and the operational endpoints
  (`/healthz`, `/readyz`, `/metrics`) only; unknown paths keep the JSON error
  envelope so API clients never receive HTML.
- **A dedicated reverse proxy owns the public origin.** Caddy
  (`web/Caddyfile`, `web/Dockerfile`) serves the built SPA and forwards
  `/api/*`, `/healthz`, `/readyz` and `/metrics` to the gateway. It uses
  `try_files {path} /index.html` for client-side routes and caches hashed
  `/assets/*` immutably.
- **Same origin, no CORS.** The proxy keeps the SPA and API on one origin, so
  the `HttpOnly`, `SameSite=lax` session cookie and the Discord OAuth callback
  keep working unchanged. This is why the SPA is not deployed to a separate
  origin/CDN.
- **Development is unchanged.** `task dev` still runs the gateway and Vite with
  HMR; Vite proxies `/api` to the gateway.

## Consequences

- The API image contains only the gateway; the SPA image contains only the
  frontend and its proxy. They build, version and deploy independently.
- Production now runs two containers (gateway + Caddy) behind one origin. TLS
  is terminated by Caddy, which also satisfies the https requirement for the
  Discord OAuth callback.
- The gateway no longer needs Node in its build; the frontend no longer needs
  Go. `task check` and `go test ./...` are unaffected by the frontend.
- The reverse proxy must not swallow unknown `/api/*` paths into the SPA
  fallback; the Caddyfile routes those to the gateway first.
