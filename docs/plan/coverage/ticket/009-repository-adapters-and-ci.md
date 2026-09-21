# Repository, adapters and CI integration

**Milestone:** C · **Gate:** G-standard, contributes to the integration gate ·
**Depends on:** 001

## Goal

Cover the handwritten repository wrappers and the database adapters through the
integration suite, and run that suite in CI with real Postgres and MariaDB.

## Context

Six `//go:build integration` files exist and skip when an env var is unset; they
are never run in CI and CI has no DB service containers. `openTestDB` is
copy-pasted three times
(`internal/plugins/store/repository/store_integration_test.go:18`,
`.../azerothaccount/repository/store_integration_test.go:18`,
`.../identitydiscord/repository/store_integration_test.go:19`). The
`postgres` (3.4%) and `azerothmysql` (6.4%) adapters and every other
`*/repository` are effectively unmeasured.

## Requirements

- A shared integration harness (build-tagged) that opens Postgres from
  `ACGW_DATABASE_URL` and MariaDB from `ACGW_AZEROTH_*_DSN`, resets tables
  between tests, and skips when unset; refactor the three duplicated
  `openTestDB` copies onto it.
- Integration tests for every handwritten repository: `adminnotes`, `apikeys`,
  `azerothaccount`, `azerothcharacter`, `identitydiscord`, `reports`, `store`,
  including error paths, pagination bounds and `LIKE`/`ILIKE` escaping.
- Adapter tests: `postgres` (pool/connection/error mapping) and `azerothmysql`
  (error paths, pagination, search escaping) in addition to the existing
  fixtures.
- Add a CI **integration** job: `postgres:17` and `mariadb:11` services with
  health checks, the MySQL init fixtures, `task db:migrate`, then
  `task test:integration`. Use ephemeral, job-local credentials; no repository
  or age secrets.
- `task coverage:integration` merges the integration profile; CI uploads the
  profile for the C1 gate.

## Tests

- The new integration tests; they skip (not fail) locally without a DB.
- A CI-only test that proves migrations apply from an empty database.

## Acceptance criteria

- The CI integration job is green on `main`.
- Each handwritten `*/repository` reaches ≥ 60% and the adapters ≥ 50%.
- No duplicated `openTestDB`; the harness is the single entry point.

## Out of scope

- Changing schema or queries beyond what tests reveal.
- Decrypting secrets in CI (see [secrets plan](../../secrets/README.md), S5).

## Notes

This ticket can land in two commits: harness + CI job first (unblocks the gate),
then per-repository tests.
