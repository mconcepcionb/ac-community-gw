# 001 — CI pipeline and baseline

**Phase:** 0 · **Gate:** G0-baseline · **Depends on:** nothing

## Goal

Establish a baseline commit and a CI pipeline that mechanically enforces every
gate in this plan, so no later ticket can regress silently.

## Findings addressed

- No CI configuration exists anywhere (`.github/workflows`, etc.).
- The OpenAPI drift check (`git diff --exit-code -- api/`) is vacuous in a
  repository with no commits: all files are untracked, so it always passes.

## Context

`task check` already chains fmt/vet/tests/architecture/OpenAPI/SPA checks, but it
is opt-in. Integration tests behind `//go:build integration` and `-race` never run.
This ticket makes the existing checks mandatory and creates the baseline the
drift check compares against.

## Atomic change

Add a GitHub Actions workflow and an initial baseline commit. No product code
changes beyond what is strictly required to make CI green (e.g. ignoring
`web/.tanstack`).

## Requirements

- Create `.github/workflows/ci.yml` with jobs:
  - **go**: install a C toolchain (`gcc`), run `task fmt:check`, `task vet`,
    `task test`, `task test:race` (or `go test -race ./...`), and
    `task openapi:check`.
  - **web**: `pnpm install --frozen-lockfile`, `task web:check`.
- Pin the Go version from `go.mod` and the pnpm version from
  `packageManager`/lockfile.
- Cache Go modules and the pnpm store.
- Make `task test` use `-count=1` so results are not served from the test cache.
- Commit the current tree as the baseline after CI is green, so
  `git diff --exit-code` in `task openapi:check` compares against a real HEAD.
- Add `web/.tanstack/` to `.gitignore` and `.dockerignore` if not already present
  (the generated router cache must not be committed or copied into images).
- Fail the workflow if the working tree is dirty after codegen (drift guard).

## Tests

- CI runs on the baseline commit and is green.
- Introduce a deliberately stale `api/swagger.yaml` in a throwaway branch and
  confirm `task openapi:check` fails; revert.
- Confirm `task test` re-runs after a no-op edit (proves `-count=1`).

## Acceptance criteria (gate G0-baseline)

- `go build ./...`, `go vet ./...`, `go test ./...` pass on a clean checkout.
- `task web:check` passes.
- CI is green on `main`.
- Baseline commit exists; `HEAD` is no longer an empty branch.
- The OpenAPI drift check demonstrably fails when the spec is stale.

## Rollback

Revert the workflow and baseline commit. No runtime behaviour is affected.

## Out of scope

- Adding the integration test job (ticket 033 adds `test:integration` and wires
  the service containers).
- Changing any application code.
