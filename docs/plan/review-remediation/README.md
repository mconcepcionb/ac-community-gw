# Code review remediation

## Status

**Delivered.** All 42 tickets have been implemented on `main`. Notable
deviations and deferred operational steps:

- **Ticket 001**: the CI workflow, `-count=1`, ignore rules and `web/.tanstack`
  entries are in place. The baseline commit is left to the operator (no commits
  were made automatically); until the repository has a commit, the
  `git diff --exit-code` drift check compares against an empty tree.
- **Ticket 010**: implemented the preferred strategy (per-request role refresh
  with a 30s cache) rather than session revocation.
- **Ticket 023**: the account/store event types had no publisher or consumer, so
  they were removed (the "remove" branch of the ticket).
- **Ticket 028**: unused generated queries were removed; `lib/pq` is retained
  because the project uses `database/sql` (pgx stdlib) and sqlc emits
  `pq.Array` for `text[]`. Replacing it requires moving to the native pgx
  interface.
- **Ticket 032**: images run non-root and have healthchecks; base-image digest
  pinning is left as an operational follow-up (requires registry access).
- **Ticket 038**: the non-envelope error rendering was fixed; URL-state
  re-sync, retry-policy centralization and the `SameSite=None` CSRF token
  remain as follow-ups.
- **Operational**: rotate the Discord client secret and SOAP password that were
  present in the local `.env`, and provide `POSTGRES_PASSWORD`,
  `MARIADB_ROOT_PASSWORD` and `MARIADB_PASSWORD` to `docker compose`.

## Goal

Close every actionable finding from the full-repository code review: exploitable
security defects first, then correctness/data-integrity bugs, then hardening,
tooling, frontend and documentation. Every fix is an **atomic ticket** and every
ticket boundary is a **gate** that must be green before the next ticket starts.

## Context

Source of this plan: the complete code review of the repository. The review
covered the Go core, plugins, adapters, the fake AzerothCore double, migrations,
sqlc, the SPA, and the build/tooling. Findings are tracked in the
[Traceability matrix](#traceability-matrix) so nothing is silently dropped.

## Scope

Everything the review found, including nits. A finding is only out of scope if
it is factually impossible to fix or explicitly accepted with a written
rationale in [Accepted deviations](#accepted-deviations).

## Principles

1. **Atomic tickets.** One ticket = one cohesive behaviour change that can be
   reviewed, tested and reverted on its own. No ticket mixes unrelated concerns.
2. **Gates between tickets.** A ticket is not complete until its gate passes.
   The standard gate applies at every ticket boundary; phase gates add heavier
   verification (integration, race, e2e, image scan) at phase boundaries.
3. **Test-first for defects.** Every bug fix lands with a test that fails before
   and passes after. Security fixes land with a negative/regression test.
4. **No wire-format changes** unless the ticket explicitly says so.
5. **Migrations are additive and reversible.** `up` and `down` are both
   exercised on a scratch database before merge.
6. **Docs move with behaviour.** Any ticket that changes observable behaviour
   updates the relevant `docs/*.md` in the same ticket.

## Gate definitions

| Gate | Applies to | Must pass |
| --- | --- | --- |
| **G-standard** | every ticket | `task check` green (fmt, vet, unit tests, architecture boundaries, OpenAPI drift, SPA checks); `task web:check` green for frontend tickets; new tests added; docs updated; no lint/coverage regression |
| **G0-baseline** | before ticket 001 | clean checkout builds; `go build ./...`, `go vet ./...`, `go test ./...` pass; baseline OpenAPI hash recorded; CI skeleton green |
| **G1-security** | end of Phase 1 | `go test -race ./...`; targeted regression tests prove each injection/authorization/leak is closed; manual penetration checklist signed off |
| **G2-correctness** | end of Phase 2 | integration tests (`-tags integration`) with Postgres + MariaDB pass; migrations `up`/`down` verified; race green |
| **G3-hardening** | end of Phase 3 | integration + negative/fuzz tests for config, body limits, middleware pass |
| **G4-data** | end of Phase 4 | integration tests pass against both databases; `task codegen:sqlc` produces no diff; OpenAPI drift green |
| **G5-infra** | end of Phase 5 | `docker compose config`/`up` smoke passes; image scan has no Critical/High; CI green on the full pipeline |
| **G6-frontend** | end of Phase 6 | `task web:check` green; Playwright smoke of login + purchase passes; no CSP violations in the browser console |
| **G7-docs** | end of Phase 7 | all doc links resolve; docs match shipped behaviour; ADRs accepted |

## Phases

| Phase | Theme | Tickets | Exit gate |
| --- | --- | --- | --- |
| 0 | Safety net and baseline | 001 | G0-baseline |
| 1 | Critical security | 002–005 | G1-security |
| 2 | High correctness and integrity | 006–010 | G2-correctness |
| 3 | Core / HTTP hardening | 011–017 | G3-hardening |
| 4 | Data and adapters | 018–028 | G4-data |
| 5 | Infrastructure and tooling | 029–033 | G5-infra |
| 6 | Frontend | 034–041 | G6-frontend |
| 7 | Documentation | 042 | G7-docs |

## Dependency graph

```
001 (CI baseline)
 │
 ├─002─003─004─005                         P1  ── G1
 │
 ├─006─007─008─009─010                     P2  ── G2
 │
 ├─011─012─013─014─015─016─017             P3  ── G3
 │
 ├─018─019─020─021─022─023─024─025─026─027─028   P4  ── G4
 │
 ├─029─030─031─032─033                     P5  ── G5
 │
 ├─034─035─036─037─038─039─040─041         P6  ── G6
 │
 └─042                                     P7  ── G7
```

Sequential order is the default. Parallelism is allowed only between tickets
that touch disjoint files and do not share a migration sequence; when in doubt,
run tickets in numeric order so each gate observes a stable tree.

Migration ordering constraint: tickets **008**, **021** and **022** are the only
ones that add migrations. They must be merged in numeric order so the single
Goose sequence stays linear.

## Ticket index

| # | Title | Phase | Depends on |
| --- | --- | --- | --- |
| 001 | [CI pipeline and baseline](ticket/001-ci-pipeline-and-baseline.md) | 0 | — |
| 002 | [Command-safe account operations](ticket/002-command-safe-account-operations.md) | 1 | 001 |
| 003 | [Command-safe admin and moderation operations](ticket/003-command-safe-admin-operations.md) | 1 | 002 |
| 004 | [Mail delivery object-level authorization](ticket/004-mail-delivery-authorization.md) | 1 | 001 |
| 005 | [Fake AzerothCore hardening](ticket/005-fake-azerothcore-hardening.md) | 1 | 001 |
| 006 | [Trusted-proxy aware rate limiting](ticket/006-trusted-proxy-rate-limiting.md) | 2 | 001 |
| 007 | [Store product ID generation](ticket/007-store-product-id.md) | 2 | 002 |
| 008 | [Order lifecycle idempotency and integrity](ticket/008-order-lifecycle-idempotency.md) | 2 | 007 |
| 009 | [OAuth state expiry and store parity](ticket/009-oauth-state-expiry.md) | 2 | 001 |
| 010 | [Authorization freshness and role revocation](ticket/010-authorization-freshness.md) | 2 | 001 |
| 011 | [Request body size limits](ticket/011-request-body-limits.md) | 3 | 001 |
| 012 | [Strict configuration parsing](ticket/012-strict-config-parsing.md) | 3 | 001 |
| 013 | [Metrics endpoint protection](ticket/013-metrics-endpoint-protection.md) | 3 | 012 |
| 014 | [Readiness redaction and timeouts](ticket/014-readiness-redaction.md) | 3 | 001 |
| 015 | [HTTP middleware and error mapping polish](ticket/015-http-middleware-polish.md) | 3 | 011 |
| 016 | [Open redirect hardening (server)](ticket/016-open-redirect-server.md) | 3 | 001 |
| 017 | [Session cookie and store hardening](ticket/017-session-cookie-store-hardening.md) | 3 | 009 |
| 018 | [MySQL adapter hardening](ticket/018-mysql-adapter-hardening.md) | 4 | 001 |
| 019 | [Upstream transport hardening (SOAP/Discord)](ticket/019-upstream-transport-hardening.md) | 4 | 012 |
| 020 | [Database integrity and constraints](ticket/020-database-integrity.md) | 4 | 008 |
| 021 | [Search and audit indexes](ticket/021-search-audit-indexes.md) | 4 | 020 |
| 022 | [Seed fixtures least privilege and idempotency](ticket/022-seed-fixtures.md) | 4 | 020 |
| 023 | [Domain events: publish or remove](ticket/023-domain-events.md) | 4 | 001 |
| 024 | [Exact-match user resolution](ticket/024-exact-match-user-resolution.md) | 4 | 001 |
| 025 | [Audit coverage for account and admin actions](ticket/025-audit-coverage.md) | 4 | 002, 003 |
| 026 | [Inactive product purchase guard](ticket/026-inactive-product-guard.md) | 4 | 007 |
| 027 | [Store delivery reconciliation](ticket/027-store-delivery-reconciliation.md) | 4 | 008 |
| 028 | [Generated SQL and type cleanup](ticket/028-generated-sql-cleanup.md) | 4 | 020 |
| 029 | [Compose exposure and secret handling](ticket/029-compose-secrets.md) | 5 | 012 |
| 030 | [Architecture test robustness](ticket/030-architecture-test.md) | 5 | 001 |
| 031 | [OpenAPI contract fixes](ticket/031-openapi-contract.md) | 5 | 001 |
| 032 | [Container hardening](ticket/032-container-hardening.md) | 5 | 029 |
| 033 | [Test and lint tooling](ticket/033-test-lint-tooling.md) | 5 | 001 |
| 034 | [Auth flow correctness and return_to](ticket/034-frontend-auth-flow.md) | 6 | 016 |
| 035 | [Product form correctness](ticket/035-frontend-product-form.md) | 6 | 001 |
| 036 | [Dialog and mutation correctness](ticket/036-frontend-dialogs-mutations.md) | 6 | 001 |
| 037 | [Route parameter and query error states](ticket/037-frontend-route-errors.md) | 6 | 001 |
| 038 | [API error and URL-state robustness](ticket/038-frontend-api-url-state.md) | 6 | 001 |
| 039 | [Caddy security headers](ticket/039-caddy-security-headers.md) | 6 | 013 |
| 040 | [Accessibility fixes](ticket/040-frontend-accessibility.md) | 6 | 001 |
| 041 | [Frontend test coverage and harness](ticket/041-frontend-test-coverage.md) | 6 | 035, 036, 037, 038 |
| 042 | [Documentation consolidation](ticket/042-documentation-consolidation.md) | 7 | all |

## Traceability matrix

Every finding from the review maps to exactly one ticket.

| Area | Finding | Ticket |
| --- | --- | --- |
| Security | Account command injection (`.account create/password/email`) | 002 |
| Security | Admin/moderation command injection (`.ban/.unban/.gmlevel`) | 003 |
| Security | Mail delivery IDOR (no ownership check) | 004 |
| Security | Fake server unauthenticated `/state` leaking passwords + wildcard CORS | 005 |
| Security | Fake server logs/journal passwords in cleartext | 005 |
| Security | Fake server default credentials, `account set password` length, mutex across DB, reset DB desync, silent ban units, body snippet leak | 005 |
| Security | Auth rate limit bypass via `X-Forwarded-For` | 006 |
| Security | OAuth state expiry ignored by Postgres store | 009 |
| Security | Session role snapshot never invalidated (revocation lag) | 010 |
| Security | Request body size limit not enforced (DoS) | 011 |
| Security | `/metrics` unauthenticated without token + Caddy proxies it | 013, 039 |
| Security | `/readyz` leaks internal DB errors | 014 |
| Security | Open redirect via backslash (`/\`) server + SPA | 016, 034 |
| Security | Upstream Discord/SOAP credentials over cleartext HTTP | 019 |
| Security | MySQL fixture grants `ALL PRIVILEGES` to `%` | 022 |
| Security | Leftover secrets in `.env`; compose interpolation echoes them | 029 |
| Security | Missing CSP / HSTS / frame-ancestors at the edge | 039 |
| Security | CSRF exposure when `SameSite=None` | 038 |
| Correctness | Product inserted with nil UUID; second create fails | 007 |
| Correctness | Order refund not idempotent; delivered order refundable | 008 |
| Correctness | `FailOrder` without `EnsureWallet`; delivered+failed-complete leaves points debited | 008, 027 |
| Correctness | Inactive products purchasable | 026 |
| Correctness | `ResolveUserByName` can miss exact matches | 024 |
| Correctness | Domain events declared but never published | 023 |
| Correctness | Missing audit records for account/admin commands | 025 |
| Correctness | `azerothmysql` requires `parseTime=true` but never enforces it | 018 |
| Correctness | Unbounded pagination offset | 018 |
| Correctness | `LIKE` wildcards from user input not escaped | 018, 021 |
| Correctness | Config parse errors silently fall back | 012 |
| Correctness | `statusRecorder` hides optional `ResponseWriter` interfaces | 015 |
| Correctness | Panics bypass access logging | 015 |
| Correctness | Client-supplied `X-Request-Id` reflected/logged unvalidated | 015 |
| Correctness | `SubscribeTyped` panics for pointer event types | 015 |
| Correctness | `StatusForError` mapping inconsistencies; `ErrNotRegistered` dead code; `RoleNames` unsorted | 015 |
| Correctness | `SetCookie` MaxAge uses wall clock; `Resolve` ignores delete error | 017 |
| Correctness | `New` does not default Audit/Readiness | 015 |
| Correctness | Readiness checks lack per-check timeout | 014 |
| Correctness | Postgres pool zero values unguarded; MySQL duplicate constructors, no idle timeout | 018 |
| Correctness | SOAP 200 without response element treated as success | 019 |
| Correctness | `HTTPClient` override silently disables timeout; body snippet in errors | 019 |
| Correctness | `goose` package globals not concurrency-safe; `MigrateSQL` duplicate | 018 |
| Data | Account link uniqueness is case-sensitive | 020 |
| Data | Wallet/order ledger destroyed by `ON DELETE CASCADE` | 020 |
| Data | Orders lack status/money constraints; wallet entry FKs/indexes | 008, 020 |
| Data | Account id type inconsistency (`integer` vs `bigint`) | 020 |
| Data | Duplicate Discord id uniqueness across tables | 020 |
| Data | Search query lacks indexes and mixes `LIKE`/`ILIKE` | 021 |
| Data | `audit_log` unbounded, no retention/partitioning | 021 |
| Data | `seed_demo_store` not idempotent; zeroed fixture credentials; admin seed grants everything | 022 |
| Data | Unused generated queries (`CreateAccountLink`, `GetAccountLinkByUsername`) | 028 |
| Data | `lib/pq` `pq.Array` dependency | 028 |
| Tooling | No CI configuration at all | 001 |
| Tooling | Integration tests never execute; no integration task | 033 |
| Tooling | OpenAPI drift check vacuous before first commit | 001 |
| Tooling | Architecture test can silently stop enforcing | 030 |
| Tooling | OpenAPI request bodies unioned with untyped object; no security scheme | 031 |
| Tooling | Swagger polish (`details: {}`, empty `externalDocs`, `required`) | 031 |
| Tooling | Compose exposes DB ports with weak credentials; app DSNs point at localhost; no restart policy | 029 |
| Tooling | Container healthchecks, non-root SPA, unpinned base images | 032 |
| Tooling | `task check` no race/`-count=1`; flaky SOAP timeout test; zero-coverage packages; no lint/coverage tooling; dotenv optional; `web/.tanstack` not ignored | 033 |
| Frontend | `RequireAuth` treats session error as anonymous (login loop) | 034 |
| Frontend | `guard.ts` dead code; `/login` no auth check; StrictMode double redirect | 034 |
| Frontend | Product update uses edited SKU as path; form never resets on prop change | 035 |
| Frontend | Confirm dialog closes on validation failure; global `invalidateQueries()`; grant dialog wrong key | 036 |
| Frontend | Invalid/zero route param yields infinite loading; wallet error shows `0`; `class=0` falsy; pagination next enabled | 037 |
| Frontend | Non-envelope errors show `[object Object]`; URL search state not re-synced; retry policy inconsistency; dead `onSubmit` | 038 |
| Frontend | Missing CSP/HSTS/frame-ancestors | 039 |
| Frontend | Tables lack `scope`/caption/labels | 040 |
| Frontend | Test coverage gaps; capture vars not reset; unused generated Zod | 041 |
| Docs | Behavioural docs drift (metrics, sessions, events, command safety, order lifecycle); secret rotation runbook | 042 |

## Risks

| Risk | Mitigation |
| --- | --- |
| Security fixes are applied inconsistently across plugins | Ticket 002 introduces one shared command-safety helper; 003 and later tickets must use it |
| Migration tickets (008, 020, 021) conflict | They are ordered by numeric ticket; each adds the next Goose version after the previous is merged |
| Refactors change the wire format | Any JSON-shape change requires an exact-JSON test in the same ticket; otherwise no wire changes |
| Large ticket count stalls delivery | Tickets are independently mergeable; phases can ship as a series of small PRs, one PR per ticket |
| `-race` unavailable locally (no cgo) | CI job in 001 installs a C toolchain so `task test:race` runs on every ticket |
| Integration tests need databases | Ticket 033 adds a `test:integration` task; CI service containers provide Postgres and MariaDB |

## Accepted deviations

None. Every review finding is in scope. If a ticket proves a finding wrong
during implementation, it records the evidence in that ticket and the
traceability matrix is updated in the same PR.

## Execution rules

1. Start at 001 and proceed in numeric order unless a ticket declares a
   dependency that allows reordering.
2. Each ticket is one PR. The PR description links the ticket.
3. Run the ticket's gate before merging. Record the gate result in the PR.
4. Never merge a ticket with a red phase gate; fix forward or revert.
5. Update this README's ticket table only to mark status; do not renumber.
