# Project skeleton and toolchain

## Goal

Create the repository structure, Go module, toolchain and development
interface.

## Context

Everything else depends on a buildable, testable project skeleton.

## Requirements

- Go module `github.com/mconcepcionb/ac-community-gw`, Go 1.27.0 minimum.
- Directory layout `cmd/`, `internal/core`, `internal/plugins`,
  `internal/adapters`, `migrations`, `docs`.
- `Taskfile.yml` with run, build, test, fmt, vet, check, codegen, db and
  docker tasks.
- `.gitignore`, `.env.example`.
- Git repository initialized with no remote assumptions.

## Acceptance criteria

- `go build ./...` succeeds.
- `task --list` shows the documented tasks.
- `task check` runs format check, vet and tests.
- No secrets are committed.

## Implementation notes

The module path is fixed. `cmd/server` is the single entrypoint. Prefer Task
over raw tool invocations.

## Tests

- `task check` passes on a clean checkout.
