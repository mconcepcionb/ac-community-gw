# 028 — Generated SQL and type cleanup

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 020

## Goal

Remove dead generated query surface and the unnecessary `lib/pq` dependency, and
make the repository types consistent.

## Findings addressed

- Nit: `CreateAccountLink` and `GetAccountLinkByUsername` are generated but
  unused; dead SQL surface increases maintenance.
- Nit: `lib/pq` is used only for `pq.Array` array encoding while the runtime
  driver is pgx; the `nonNilRoles` workaround exists because of it.

## Context

`internal/plugins/azerothaccount/repository/queries.sql`,
`internal/plugins/identitydiscord/repository/{queries.sql,store.go}`, generated
code, `sqlc.yaml`, `go.mod`.

## Atomic change

Delete unused queries, switch array encoding to a pgx-native approach (or keep
`lib/pq` with a written rationale), regenerate sqlc, and tidy modules.

## Requirements

- Remove `CreateAccountLink`/`GetAccountLinkByUsername` if truly unused (verify
  with `rg` across Go and docs); regenerate.
- Replace `pq.Array` encoding with a pgx-compatible type (e.g. `[]string` handled
  by sqlc/pgx, or `pgtype`) and remove the `nonNilRoles` workaround if it becomes
  unnecessary; if `lib/pq` must stay, document why.
- Run `go mod tidy` and regenerate the OpenAPI spec if any DTO changed.
- Ensure `task codegen:sqlc` produces no diff afterward.

## Tests

- `go build ./...` and all repository tests pass after regeneration.
- `task codegen:sqlc` is clean (no diff).
- Roles round-trip test (nil and non-nil) still passes.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green; `task openapi:check` green.
- No unused generated query remains.
- `lib/pq` is either removed or documented as intentional.

## Rollback

Revert the regeneration and module changes.

## Out of scope

- Adding new queries except those required by ticket 024.
