# 030 — Architecture test robustness

**Phase:** 5 · **Gate:** G-standard, contributes to G5-infra · **Depends on:** 001

## Goal

Make the import-boundary test impossible to silently disable and cover the whole
module, not just `internal/`.

## Findings addressed

- Medium: `internal/architecture/architecture_test.go` hardcodes the module path,
  walks only `internal/`, skips anything outside `internal/core|adapters|plugins`,
  and never asserts that it inspected any files. A module rename or a new
  top-level package (e.g. `pkg/`, a relocated plugin) passes while checking
  nothing.

## Context

`docs/architecture.md` promises the rules are mechanically enforced. The current
test can be a no-op without failing.

## Atomic change

Derive the module path, analyze the full package graph, and fail when coverage of
the inspection is empty.

## Requirements

- Derive `modulePath` from `go.mod` (or `go list -m` / `debug.ReadBuildInfo`), not
  a literal.
- Walk all module packages (`go/packages` or walk the module root), classifying
  every package, and enforce:
  - core must not import plugins or adapters;
  - plugins must not import adapters or sibling plugins;
  - adapters must not import plugins;
  - core/plugins must not import adapters (per the documented rules).
- Assert that at least one file/import was inspected; `t.Fatal` otherwise.
- Cover new top-level directories by failing on unknown internal packages unless
  they are explicitly classified.
- Keep `TestSourceIsFormatted` and extend it to cover `.go` files outside
  `internal/` (already walks the root; verify skip list).

## Tests

- Test: the boundary test fails when a synthetic forbidden import is added
  (table/fixture-based, or a unit test of the classifier).
- Test: the module path is read correctly after a simulated rename.
- Test: an unknown top-level package fails classification.

## Acceptance criteria (gate G-standard, G5-infra)

- `task check` green.
- The test fails if it inspects zero packages.
- A forbidden import in any top-level package fails the test.

## Rollback

Revert; enforcement strength returns to the old level.

## Out of scope

- Runtime dependency-injection checks.
