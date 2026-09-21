# Coverage tooling, measured set and gate

**Milestone:** A · **Gate:** G-standard · **Depends on:** —

## Goal

Produce a cross-platform coverage report over an explicit **measured set**
(handwritten code, generated code excluded) and enforce a ratcheting floor in
CI.

## Context

`task coverage` runs `go test -count=1 -coverprofile=coverage.txt ./...` then
`go tool cover -func=coverage.txt` (`Taskfile.yml:58-62`). The default profile
includes the `repository/generated` packages at 0%, which drags the global
figure and hides the real signal. Developers use Windows, so the filter cannot
rely on `grep`/`awk`.

## Requirements

- Add `cmd/covercheck` (pure Go, cross-platform):
  - flags: repeated `-profile` (merged), `-ignore coverage.ignore`,
    `-floor coverage.floor`, optional `-json`;
  - prints a per-package table and the total for the measured set;
  - exits non-zero when the total is below the floor.
- Add `coverage.ignore` with the excluded prefixes:
  `internal/plugins/*/repository/generated`, `internal/architecture`,
  `internal/fake/azerothcore`, `cmd/`.
- Add `coverage.floor` with the current measured total (the baseline).
- Replace `task coverage` with: unit profile + `covercheck` table. Add
  `task coverage:integration` (unit + integration profiles) and
  `task coverage:check` (enforce the floor).
- Add `task coverage:check` to the Go CI job.
- `coverage.txt` and any summary output stay gitignored (already covered).

## Tests

- `cmd/covercheck` unit tests: profile parsing, per-package aggregation, ignore
  matching, multiple-profile merge, floor pass/fail exit codes.
- A fixture profile that contains a generated package proves it is excluded
  from the denominator.

## Acceptance criteria

- `task coverage:check` is green at the recorded floor on a clean checkout.
- `repository/generated` (and the other ignored prefixes) are absent from the
  denominator and present in the printed table as excluded.
- The Go CI job fails when the floor is raised above the achieved total.

## Out of scope

- Raising coverage (C3-C10); C1 only measures and gates.
- Changing test execution semantics.

## Rollback

Remove the CI step and the `coverage:*` tasks; `task coverage` can revert to the
`go tool cover` command. `cmd/covercheck` is additive.
