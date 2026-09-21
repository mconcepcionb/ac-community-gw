# Go linter

**Milestone:** A · **Gate:** G-standard · **Depends on:** 001

## Goal

Add `golangci-lint` with a curated rule set to the Go gate, without unrelated
refactors.

## Context

There is no Go linter config anywhere (`golangci*`, `revive.toml`,
`staticcheck.conf` all absent); the Go gate is `gofmt` + `go vet` only. The SPA
already has Biome. `golangci-lint` is the documented recommendation in the
[status report](../../../status-report.md).

## Requirements

- Add `.golangci.yml` (version pinned) enabling at least: `govet`,
  `staticcheck`, `errcheck`, `ineffassign`, `unused`, `gofmt`/`goimports`, and a
  small `revive` subset. Keep the set stable and justified.
- Exclude generated paths (`internal/plugins/*/repository/generated`) and
  `internal/fake/azerothcore` if they produce noise.
- Triage pre-existing findings: fix the trivial ones and record any
  intentional `//nolint` with a reason. No behaviour changes.
- Add `task lint` and a CI step after `task vet`.
- Pin the `golangci-lint` version in CI (the action or a `go install` pin).

## Tests

- `task lint` exits clean.
- The CI step fails on a deliberately introduced lint error.

## Acceptance criteria

- `task lint` and the CI step are green.
- No production behaviour changed; any `nolint` carries a reason.
- The config and the pinned version are documented in `docs/development.md`.

## Risks

- The newest `golangci-lint` may not yet support Go 1.27; pin a compatible
  released version and keep the rule set small if a rule is unsupported.

## Out of scope

- Adding coverage rules or thresholds (C1).
- Style changes unrelated to the enabled linters.
