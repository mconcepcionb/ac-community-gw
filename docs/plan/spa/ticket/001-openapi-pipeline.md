# OpenAPI pipeline

## Goal

Introduce the OpenAPI generation pipeline so the gateway contract can be
documented incrementally and consumed by the frontend codegen.

## Context

There is no machine-readable contract today. The frontend needs one. `swag v2`
converts Go annotations into an OpenAPI document without changing the
`net/http` style of the codebase.

## Requirements

- Add `github.com/swaggo/swag/v2/cmd/swag` as a Go tool (`go get -tool`).
- General API info annotation block in `cmd/server/main.go` (`@title`,
  `@version`, `@description`, `@BasePath /api/v1`, `@schemes`).
- Annotate the operational endpoints `GET /healthz` and `GET /readyz` and the
  shared `httpapi.ErrorResponse` / `httpapi.ErrorBody` schemas.
- Emit `api/swagger.yaml` (OpenAPI 3.1) with `--ot yaml`, no generated Go.
- Taskfile: `openapi:spec` (run swag) and `openapi:check:spec` (regenerate and
  fail on drift via `git diff --exit-code api/`).
- Document the annotation conventions (tags, stable `@ID`, DTO references) in
  this plan's README.

## Acceptance criteria

- `task openapi:spec` writes `api/swagger.yaml`.
- The document validates as OpenAPI 3.1 and contains `/healthz`, `/readyz` and
  the error envelope schema.
- `task openapi:check:spec` is green immediately after generation and red after
  a manual edit.
- `task check` is unaffected and green.

## Implementation notes

- swag is only a developer tool; no runtime dependency and no import of
  generated `docs.go` (output is YAML only).
- `swag v2` has no OpenAPI 3.0 mode; the spec is emitted as OpenAPI 3.1 with
  `--v3.1`. If a 3.1 generator defect appears, the fallback is the default
  Swagger 2.0 output (also consumable by hey-api).
- Parse `./cmd/server,./internal` with `--parseInternal`; the general info file
  is `main.go` in the first `-d` directory. swag logs a harmless
  "no Go files in .../internal" warning because the directory has no package of
  its own.
- Full paths are used in `@Router` (no `@BasePath`) so operational endpoints
  outside `/api/v1` are correct; `@accept`/`@produce` are omitted (deprecated at
  the top level in 3.1).
- Schema names are set with a trailing `@name` comment on the type declaration
  (`} // @name ErrorResponse`); the leading doc comment is used as description
  and is not read for `@name`.
- The client tasks (`openapi:client`, `openapi`, `openapi:check`) are added in
  ticket 020, once `web/` exists, to keep every task runnable.
- `openapi:check:spec` uses `git diff --exit-code`, so `api/swagger.yaml` must be
  committed for the check to have teeth.

## Tests

- None in Go: the artifact is the spec. The drift check is the guard.
- Manually confirm with `go tool swag init --v3.1 -g main.go
  -d ./cmd/server,./internal --parseInternal -o ./api --ot yaml`.

## Dependencies

- None. Unblocks tickets 002..018 and 020.
