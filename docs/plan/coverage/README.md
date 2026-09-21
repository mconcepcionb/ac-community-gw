# Test coverage

## Goal

Raise the measured coverage of the handwritten Go code (and the SPA) to a
meaningful floor, stop counting generated code against the denominator, add a
Go linter, and wire a **ratchet** into CI so coverage only goes up.

## Status

| Area | State |
| --- | --- |
| Baseline measured (2026-09-21) | done |
| Coverage tooling, measured set and gate | implemented (C1) |
| `golangci-lint` in the Go gate | implemented (C2) |
| Shared test support + core packages | implemented (C3) |
| Plugin domain packages | implemented (C4-C7): apikeys 81%, reports 81%, store 82%, azerothaccount 82%, azerothcharacter 77%, identitydiscord 81%, adminnotes 87% |
| Plugin registration convergence | implemented (C8) |
| Repository/adapters + CI integration job | implemented (C9); repositories 73-87%, postgres 94%, azerothmysql 58% |
| Frontend coverage thresholds | planned (C10) |
| Ratchet and documentation | planned (C11) |

## Decisions

- **Global floor: 60%** of the measured set (handwritten Go, excluding
  generated/wiring-only packages), enforced by a ratchet (C1 sets the baseline,
  C11 raises it to the achieved total).
- **`golangci-lint` with a curated config** is added to the Go gate (C2).
- **CI gains an integration job** with Postgres + MariaDB service containers
  (C9); without it repository/adapter coverage cannot be verified.
- **SPA: 60% statements**, plumbing and thresholds first (C10); behavioural gaps
  stay in `review-remediation/041`.
- **CI stays secret-free**; the integration job uses ephemeral credentials.

## Context

The one-off [status report](../../status-report.md) measured 36.6% global
coverage. A fresh run on 2026-09-21 (`go test -count=1 -cover ./...`) shows the
distribution is **bimodal**: core infrastructure is well covered and the domain
plugins are not. Generated `repository/generated` packages report 0% and drag the
global figure, which hides the real signal.

| Package | Coverage |
| --- | --- |
| `internal/core/metrics` | 100.0% |
| `internal/core/azerothcore` | 93.3% |
| `internal/core/events` | 90.0% |
| `internal/core/config` | 85.1% |
| `internal/core/persistence` | 84.2% |
| `internal/core/commands` | 81.0% |
| `internal/adapters/azerothsoap` | 80.6% |
| `internal/core/httpapi` | 78.2% |
| `internal/adapters/discord` | 75.9% |
| `internal/core/permissions` | 75.4% |
| `internal/fake/azerothcore` | 72.7% |
| `internal/core/auth` | 62.2% |
| `internal/core/services` | 62.2% |
| `internal/plugins/gatewayadmin` | 60.7% |
| `internal/plugins/azerothitem` | 54.2% |
| `internal/plugins/azerothadmin` | 51.7% |
| `internal/plugins/azerothinfo` | 50.9% |
| `internal/plugins/identitydiscord` | 50.7% |
| `internal/core/logging` | 45.5% |
| `internal/core/itemview` | 43.5% |
| `internal/plugins/store` | 38.2% |
| `internal/plugins/azerothaccount` | 36.4% |
| `internal/plugins/azerothcharacter` | 36.1% |
| `internal/plugins/adminnotes` | 35.7% |
| `internal/plugins/reports` | 24.4% |
| `internal/adapters/azerothmysql` | 6.4% |
| `internal/adapters/postgres` | 3.4% |
| `internal/core/audit` | 0.0% |
| `internal/core/azerothdb` | 0.0% |
| `internal/core/ttlcache` | 0.0% |
| `internal/core/plugins` | 0.0% |
| `internal/plugins/apikeys` | 0.0% |
| `internal/plugins/*/domain` | 0.0% |
| `internal/plugins/*/repository` (+`generated`) | 0.0% |
| `cmd/*` | 0.0% |

### Findings that shape the work

- **No shared test harness.** Every plugin declares its own unexported fakes and
  calls unexported handler methods directly with `httptest`
  (`internal/plugins/store/handlers_test.go`, `.../reports/handlers_test.go`,
  `.../azerothadmin/http_test.go`). A small shared support package should remove
  the repeated request/decode/auth boilerplate; fakes stay per package.
- **`Register()` and `permissionDefs` are untested in every plugin.** No test
  builds `plugins.Registry` or calls `Manager.RegisterAll`
  (`internal/core/plugins/registry.go:32`), so route/permission wiring is
  unverified (C8).
- **Integration infrastructure is duplicated and absent from CI.** Six
  `//go:build integration` files, `openTestDB` copy-pasted three times, all skip
  on a missing env var; CI has no DB services and never runs
  `task test:integration` (C9).
- **The profile includes generated code.** The default `coverage.txt` contains
  the 0% `repository/generated` packages, so a cross-platform filter is needed
  (C1 ships a small Go checker rather than `grep`/`awk`, since developers use
  Windows).
- **The SPA has no coverage provider.** `@vitest/coverage-v8` is absent from
  `web/package.json` and the lockfile; there is no `coverage` block in
  `web/vite.config.ts` and no `test:coverage` script (C10).

## Scope

- Tooling: a coverage task that excludes generated/wiring-only packages and
  prints a per-package table; a threshold check; `golangci-lint`.
- Backend tests for the low-coverage **domain and handler** packages, test-first
  where a defect is known and behaviour-first otherwise.
- A shared, non-plugin test support package.
- Repository/adapters covered through the existing `-tags integration` suite,
  run in a new CI job with service containers.
- Frontend coverage plumbing and thresholds for `web/src`.
- CI wiring and a documented floor that ratchets.

## Out of scope

- A percentage target for its own sake. A package may stay low if its uncovered
  lines are wiring that no test can meaningfully exercise *and* it is excluded
  from the denominator.
- Covering `cmd/*` boots (`cmd/server` wiring, `openapi-postprocess`,
  `fakeazerothcore`) beyond what is cheaply reachable.
- Rewriting well-covered core packages.
- Reimplementing the behavioural frontend gaps already listed in
  `review-remediation/041`; C10 owns the plumbing and thresholds.

## Principles

1. **Coverage is a compass, not a target.** It points at untested behaviour; the
   deliverable is the test, not the number.
2. **Test at the seam.** Handlers with `httptest` and fake services; domain
   logic directly; repositories with integration tests against Postgres/MariaDB.
3. **Generated and wiring-only code is excluded from the denominator**, never
   from the report.
4. **Ratchet.** The gate fails when coverage of the measured set drops below the
   recorded floor. The floor rises when a ticket raises it.
5. **Integration tests count.** The `-tags integration` suite runs in CI; its
   profile is merged into the measured total.

## Design

### Measurement set

`./internal/...` excluding:

- `internal/.../repository/generated`
- `internal/architecture` (test-only, no statements)
- `internal/fake/azerothcore` (test double)

`cmd/*` is reported but excluded from the floor. The excluded prefixes live in a
checked-in `coverage.ignore` file read by the tooling (C1), so the numerator and
denominator cannot drift.

### Targets

Targets are floors for the floor-raising tickets; they are deliberately below
"100%" to reward behavioural tests, not line touching.

| Group | Packages | Target |
| --- | --- | --- |
| Core gaps | `core/audit`, `core/azerothdb`, `core/ttlcache`, `core/plugins`, `core/logging`, `core/itemview` | ≥ 70% |
| Domain plugins | `reports`, `azerothaccount`, `azerothcharacter`, `store`, `adminnotes`, `identitydiscord`, `apikeys` | ≥ 60% |
| Repository layer | each `*/repository` (handwritten wrapper) | ≥ 60% via integration |
| Adapters | `postgres`, `azerothmysql` | ≥ 50% via integration |
| Measured set (global) | all of the above | floor set by C1, 60% by C11 |
| SPA (`web/src`, excluding generated + `routeTree.gen.ts`) | — | ≥ 60% statements |

### Tooling

- `cmd/covercheck` (Go, cross-platform) reads one or more coverage profiles,
  applies `coverage.ignore`, prints a per-package table and a total, and with
  `-floor coverage.floor` fails on regression.
- `task coverage` runs unit tests with `-coverprofile`, then `covercheck` for
  the table. `task coverage:integration` adds the integration-tagged profile.
- `task coverage:check` runs both and enforces the floor; it is added to the Go
  CI job.
- `task lint` runs `golangci-lint run` with `.golangci.yml`; CI runs it in the
  Go job.
- Frontend: Vitest coverage (`@vitest/coverage-v8`) with `thresholds` in
  `web/vite.config.ts`, a `test:coverage` script and a `task web:coverage:check`.

## Backend gaps

| Gap | Needed by |
| --- | --- |
| `cmd/covercheck`, `coverage.ignore`, `coverage.floor` | C1 |
| `.golangci.yml` and `task lint` | C2 |
| `internal/testsupport` (request/decode/auth helpers, no plugin imports) | C3 |
| CI integration job with Postgres + MariaDB services | C9 |
| Vitest coverage provider and thresholds | C10 |

## Milestones

| Milestone | Tickets |
| --- | --- |
| A - Tooling and gate | C1, C2 |
| B - Support, core and plugins | C3, C4, C5, C6, C7, C8 |
| C - Repository, adapters and CI | C9 |
| D - Frontend and ratchet | C10, C11 |

### Dependency graph

```
C1 -> C2, C3, C4, C5, C6, C7, C8, C9, C10
C3 -> C4, C5, C6, C7, C8
C4..C8 -> C11
C9 -> C11
C10 -> C11
```

C1 defines the measured set and the floor everything else is judged against; C3
introduces the shared support the plugin tickets reuse. C11 runs last so the
floor is raised once.

## Risks

| Risk | Mitigation |
| --- | --- |
| Chasing the number produces brittle tests | Principles 1-2: test behaviour at seams; review each test for signal |
| Generated code keeps dragging the global figure | C1 excludes it from the denominator in the tooling and CI |
| Integration tests never run locally | C9 uses the same service containers as CI and `task test:integration` |
| Coverage floor blocks unrelated work | The floor only fails on regression; fix forward by adding tests |
| `golangci-lint` floods the build with pre-existing findings | C2 lands a curated linter set with a triaged baseline, no unrelated refactors |
| Newest golangci-lint does not yet support Go 1.27 | C2 pins the version in CI and the config; keep the rule set small and stable |
| `internal/testsupport` becomes a god package / crosses boundaries | Helpers only (stdlib + `core/auth`), never imports a plugin; enforced by the architecture test |

## Tickets

### Milestone A - Tooling and gate

1. [001-coverage-tooling-and-gate.md](ticket/001-coverage-tooling-and-gate.md) - `covercheck`, measured set, floor, CI step
2. [002-go-linter.md](ticket/002-go-linter.md) - curated `golangci-lint` config, task and CI step

### Milestone B - Support, core and plugins

3. [003-test-support-and-core.md](ticket/003-test-support-and-core.md) - `internal/testsupport` + core packages at 0%
4. [004-apikeys-and-reports.md](ticket/004-apikeys-and-reports.md) - the two lowest-covered plugins
5. [005-store.md](ticket/005-store.md) - refund/retry, wallet/order service, product CRUD
6. [006-azerothaccount-and-character.md](ticket/006-azerothaccount-and-character.md) - claims, self-service, leaderboards, visibility
7. [007-identitydiscord-and-adminnotes.md](ticket/007-identitydiscord-and-adminnotes.md) - roles admin, provisioning, notes delete/audit
8. [008-plugin-registration.md](ticket/008-plugin-registration.md) - `Register`/`permissionDefs` convergence across plugins

### Milestone C - Repository, adapters and CI

9. [009-repository-adapters-and-ci.md](ticket/009-repository-adapters-and-ci.md) - shared integration harness, repository/adapter tests, CI integration job

### Milestone D - Frontend and ratchet

10. [010-frontend-coverage.md](ticket/010-frontend-coverage.md) - coverage provider, thresholds, high-value gaps
11. [011-ratchet-and-docs.md](ticket/011-ratchet-and-docs.md) - raise the floor, document the measurement set
