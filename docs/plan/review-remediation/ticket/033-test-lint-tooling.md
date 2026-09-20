# 033 — Test and lint tooling

**Phase:** 5 · **Gate:** G-standard, contributes to G5-infra · **Depends on:** 001

## Goal

Make the test suite actually run integration coverage, add race/lint tooling,
remove flakiness, raise coverage on untested packages, and make the Taskfile
robust.

## Findings addressed

- Medium: six integration test files behind `//go:build integration` never run;
  no `test:integration` task exists.
- Low: `task check` uses cached tests and no race run; `test`/`test:race` lack
  `-count=1`.
- Low: the SOAP timeout test uses a 30 ms timeout against a 200 ms handler sleep
  (flaky on loaded CI).
- Low: zero-coverage packages with real logic (`adapters/postgres`,
  `core/persistence`, `core/plugins`, `core/audit`, `core/logging`).
- Low: no coverage task, no `staticcheck`/`.golangci.yml`, no `.editorconfig`.
- Nit: `task` dotenv is silently optional, so a missing `.env` fails cryptically
  in `db:migrate`.
- Nit: `web/.tanstack/` is neither gitignored nor dockerignored (also handled in
  001; verify).

## Context

`Taskfile.yml`, `docs/development.md`, test files.

## Atomic change

Add the missing tasks/tooling and tests; no application behaviour changes.

## Requirements

- Add `test:integration` running `go test -tags integration ./...`; document how
  to start the required databases locally.
- Add CI service containers (Postgres + MariaDB) and run `test:integration` on
  every PR.
- Add `-count=1` to `test`/`test:race`; include `test:race` in `check`.
- Add a `coverage` task and a minimum threshold or a documented baseline.
- Add `staticcheck` (or golangci-lint) and `.editorconfig`; wire lint into
  `check`.
- Fix the flaky SOAP test (larger margins or an injectable transport/clock).
- Add unit tests for the zero-coverage packages (readiness registry, postgres
  empty-URL/error paths, plugin/audit registries, logging level parsing).
- Add a `test -f .env` preflight guard (or clear error) to DB tasks.
- Ensure `web/.tanstack/` is ignored in `.gitignore` and `.dockerignore`.

## Tests

- `task check` and `task test:integration` pass with databases up.
- The flaky test passes repeatedly (`-count=10`).
- Coverage report shows the previously-zero packages at >0.

## Acceptance criteria (gate G-standard, G5-infra)

- CI runs unit, race and integration suites.
- `task check` includes race and lint.
- No test depends on tight wall-clock margins.

## Rollback

Revert tasks/tests; application behaviour is unaffected.

## Out of scope

- Changing application logic to make it testable (handled per feature ticket).
