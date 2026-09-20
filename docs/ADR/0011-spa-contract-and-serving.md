# ADR 0011: SPA contract, sessions and serving

## Status

Accepted. The **serving** decision (embedded SPA) is superseded by
[ADR 0012](0012-decoupled-spa-serving.md); the contract and session decisions
remain in force.

## Context

The gateway had a throwaway `web/index.html` test frontend. We want a
production-grade single-page application (React + TypeScript) while keeping the
backend a `net/http` modular monolith with a single binary. Two questions had to
be answered:

1. How is the HTTP contract shared with the frontend without hand-maintained
   duplicate types?
2. How is the SPA served, given that config validation forbids `ACGW_WEB_DIR`
   in production?

## Decision

- **Contract from Go annotations.** Handlers are annotated with `swag v2`
  comments; `task openapi` generates `api/swagger.yaml` (OpenAPI 3.1) and, with
  `@hey-api/openapi-ts`, a typed TypeScript client (SDK, React Query options,
  Zod schemas). Both are committed and guarded by `task openapi:check`.
  `operationId`s are stable via `@ID`; schema names use a trailing `@name`
  comment.
- **Server-side cookie sessions for the SPA.** The SPA never stores tokens. It
  resolves the principal from `GET /api/v1/me` (401 = anonymous) and relies on
  the `HttpOnly` cookie, with `credentials: "include"` on the generated client.
  `/me` also returns the effective permissions so the UI can be
  permission-aware.
- **Embedded SPA in production.** `internal/core/webui` embeds the built assets
  with `//go:embed all:dist` under the `embedui` build tag; `httpapi` serves an
  `fs.FS` with an `index.html` fallback for client routes. `task build:spa` and
  the Docker multi-stage build produce a single self-contained binary.
  `ACGW_WEB_DIR` remains a development-only override.

## Consequences

- The contract cannot drift silently: annotations and generated artifacts are
  checked together in `task check`.
- Go-only builds and tests do not require the frontend; the `embedui` tag keeps
  the embed out of ordinary builds.
- Production is one artifact (the binary), matching the single-binary design;
  there is no separate static host to deploy.
- Adding an endpoint requires an annotation and a `task openapi` run; forgetting
  it fails `task check`.
