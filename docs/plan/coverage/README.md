# Test coverage

## Goal

Raise the measured coverage of the handwritten Go code (and the SPA) to a
meaningful floor, stop counting generated code against the denominator, add a
Go linter, and wire a **ratchet** into CI so coverage only goes up.

## Status

| Area | State |
| --- | --- |
| Baseline measured (2026-09-21) | done |
| Coverage tooling and generated-code exclusion | planned |
| `golangci-lint` in the Go gate | planned |
| Core packages at 0% | planned |
| Plugin domain packages | planned |
| Repository/adapters via integration tests | planned |
| Frontend coverage thresholds | planned |
| Ratchet gate in CI | planned |

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

## Scope

- Tooling: a coverage task that excludes generated/wiring-only packages and
  prints a per-package table; a threshold check; `golangci-lint`.
- Backend tests for the low-coverage **domain and handler** packages, test-first
  where a defect is known and behaviour-first otherwise.
- Repository/adapters covered through the existing `-tags integration` suite
  (the code path that needs a real database).
- Frontend coverage thresholds for `web/src` with Vitest.
- CI wiring and a documented floor that ratchets.

## Out of scope

- A percentage target for its own sake. A package may stay low if its uncovered
  lines are wiring that no test can meaningfully exercise *and* it is excluded
  from the denominator.
- Covering `cmd/*` boots (server wiring, `openapi-postprocess`,
  `fakeazerothcore`) beyond a smoke test.
- Rewriting well-covered core packages.

## Principles

1. **Coverage is a compass, not a target.** It points at untested behaviour; the
   deliverable is the test, not the number.
2. **Test at the seam.** Handlers with `httptest` and fake services; domain
   logic directly; repositories with integration tests against Postgres/MariaDB.
3. **Generated and wiring-only code is excluded from the denominator**, never
   from the report.
4. **Ratchet.** The gate fails when coverage of the measured set drops below the
   recorded floor. The floor rises when a ticket raises it.
5. **Integration tests count.** The `-tags integration` suite runs in CI with
   service containers; its coverage is merged with the unit profile.

## Design

### Measurement set

`internal/...` excluding:

- `internal/.../repository/generated`
- `internal/architecture` (test-only, no statements)
- `internal/fake/azerothcore` (test double)

`cmd/*` is reported but excluded from the floor. The measured set is defined in
one place (a `coverage:ignore` list read by the tooling task and CI) so the
numerator and denominator cannot drift.

### Targets

Targets are floors for the floor-raising tickets; they are deliberately below
"100%" to reward behavioural tests, not line touching.

| Group | Packages | Target |
| --- | --- | --- |
| Core gaps | `core/audit`, `core/azerothdb`, `core/ttlcache`, `core/plugins`, `core/logging`, `core/itemview` | ≥ 70% |
| Domain plugins | `reports`, `azerothaccount`, `azerothcharacter`, `store`, `adminnotes`, `identitydiscord`, `apikeys` | ≥ 60% |
| Repository layer | each `*/repository` (handwritten wrapper) | ≥ 60% via integration |
| Adapters | `postgres`, `azerothmysql` | ≥ 50% via integration |
| Measured set (global) | all of the above | floor set by ticket C1, ≥ 60% by C7 |
| SPA (`web/src`, excluding generated + `routes/*.gen.ts`) | — | ≥ 60% statements |

### Tooling

- `task coverage` gains `-coverpkg` over the measured set, a per-package summary
  and a machine-readable total.
- `task coverage:check` compares the total against `coverage.floor` (checked in)
  and fails on regression.
- `task lint` runs `golangci-lint run`; CI runs it in the Go job.
- Frontend: Vitest `--coverage` with `thresholds` in `web/vite.config.ts` (or a
  dedicated coverage config) and a `task web:coverage:check`.

## Backend gaps

| Gap | Needed by |
| --- | --- |
| Coverage-ignore list + measured-set task | C1 |
| `golangci-lint` config and task | C2 |
| Fake services/fixtures reusable across handler tests | C3, C4 |
| Integration harness that seeds Postgres + MariaDB | C5 |
| Frontend coverage provider wired | C6 |

## Milestones

| Milestone | Tickets |
| --- | --- |
| A - Tooling and gate | C1, C2 |
| B - Core and plugins | C3, C4 |
| C - Repository and adapters | C5 |
| D - Frontend and ratchet | C6, C7 |

### Dependency graph

```
C1 -> C2, C3, C4, C5, C6
C3, C4, C5 -> C7
C6 -> C7
```

C1 defines the measured set and the floor everything else is judged against.
C2 adds the linter independently. C7 only after the test tickets land, so the
floor is raised once.

## Risks

| Risk | Mitigation |
| --- | --- |
| Chasing the number produces brittle tests | Principles 1-2: test behaviour at seams; review each test for signal |
| Generated code keeps dragging the global figure | C1 excludes it from the denominator in the tooling and CI |
| Integration tests never run locally | C5 uses the same service containers as CI and `task test:integration` |
| Coverage floor blocks unrelated work | The floor only fails on regression; fix forward by adding tests |
| `golangci-lint` floods the build with pre-existing findings | C2 lands with a curated linter set and a triaged baseline, no unrelated refactors |

## Tickets

### Milestone A - Tooling and gate

- **C1 - Coverage tooling and measured set.** Add the coverage-ignore list and a
  `task coverage` that reports the measured set per package and writes the total
  to `coverage.txt`; add `task coverage:check` with a checked-in
  `coverage.floor`. Record the baseline. Wire the check into the Go CI job.
  *Gate:* `task coverage:check` passes at the recorded floor; generated packages
  are absent from the denominator.

- **C2 - Go linter.** Add `.golangci.yml` (curated: `govet`, `staticcheck`,
  `errcheck`, `ineffassign`, `unused`, `gofmt`/`goimports`, `revive` subset),
  a `task lint`, and a CI step. Triage or explicitly disable rules that conflict
  with documented decisions. *Gate:* `task lint` clean; no behaviour change.

### Milestone B - Core and plugins

- **C3 - Core packages at 0%.** Tests for `core/ttlcache`, `core/plugins`
  (registry), `core/audit` (recorder contract, redaction of sensitive fields),
  `core/azerothdb` (row→domain mapping with a fake DB), `core/logging` and
  `core/itemview`. *Gate:* each reaches its target; no production change without
  a test-first defect.

- **C4 - Domain plugins.** Behavioural tests per ticket-sized slice for
  `reports` (submission, lifecycle, queue read), `store` (purchase, idempotency,
  refund/retry, reconciliation), `azerothaccount` and `azerothcharacter`
  (authorization, ownership, admin actions), `adminnotes`, `identitydiscord`
  (role mapping, permission matrix) and `apikeys` (scopes, auth). Reuse the
  handler test harness from C3. *Gate:* each reaches its target; every fix has a
  regression test.

### Milestone C - Repository and adapters

- **C5 - Repository/adapters via integration.** A seeding harness for Postgres
  (goose + fixtures) and MariaDB (the fake AC fixture); integration tests for
  each handwritten `*/repository` and both adapters, including error paths,
  pagination bounds and `LIKE` escaping. Merge integration coverage into the
  measured profile. *Gate:* targets met; `task test:integration` green in CI.

### Milestone D - Frontend and ratchet

- **C6 - Frontend coverage.** Enable Vitest coverage for `web/src` excluding
  generated client and `routeTree.gen.ts`; add thresholds and
  `task web:coverage:check`; fill the highest-value gaps (auth flow, mutations,
  error states). *Gate:* `task web:check` and the new check green.

- **C7 - Ratchet and documentation.** Raise `coverage.floor` to the achieved
  total, document the measurement set, the floor and how to raise it in
  `docs/development.md`, and add the coverage step to the CI gate description.
  *Gate:* floor ≥ 60% on the measured set; docs updated.
