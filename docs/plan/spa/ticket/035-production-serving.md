# Production serving

> Superseded by [ADR 0012](../../../ADR/0012-decoupled-spa-serving.md). The
> gateway no longer embeds the SPA; a dedicated Caddy reverse proxy serves it.

## Goal

Ship the SPA embedded in the gateway binary as a single self-contained
artifact.

## Context

The gateway is a single-binary modular monolith, and config validation forbids
`ACGW_WEB_DIR` when `ACGW_ENV=production`. The SPA therefore cannot be served
from a directory in production; it is compiled into the binary instead.

## Requirements

- `internal/core/webui` exposes `Assets() fs.FS`, populated when the binary is
  built with the `embedui` tag (`//go:embed all:dist`). Without the tag it
  returns nil and the gateway falls back to `ACGW_WEB_DIR` (development only).
- `httpapi.Dependencies.WebFS` takes precedence over `Config.Web.Dir`; the SPA
  handler serves assets from the filesystem and falls back to `index.html` for
  extension-less paths, keeping the JSON envelope for missing assets and
  `/api/*`.
- `task build:spa`: runs `web:build`, copies `web/dist` to
  `internal/core/webui/dist`, and builds with `-tags embedui`.
- Multi-stage Dockerfile: a Node stage runs `pnpm install --frozen-lockfile` and
  `pnpm build`; the Go stage copies the resulting `dist` into the embed path and
  builds with `-tags embedui`. The final image stays distroless/nonroot.
- `web/dist` and `internal/core/webui/dist` are git-ignored.

## Acceptance criteria

- `task build:spa` produces a binary that serves the SPA at `/` and deep client
  routes, with `/api/v1/unknown` returning the JSON 404 envelope.
- `task docker:build` produces an image with the same behaviour.
- `task check` and `task web:check` stay green. `go build ./...` (no tag) and
  `go test ./...` do not require the frontend build.

## Implementation notes

- Embedding requires the assets to exist at compile time, hence the copy step.
  The `embedui` tag keeps ordinary Go builds and tests free of that
  requirement.
- The production config guard is unchanged: `ACGW_WEB_DIR` remains
  development-only.

## Dependencies

- 001..034 (at least the features to ship), 003, 019.
