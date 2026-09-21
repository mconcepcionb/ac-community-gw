# apikeys and reports

**Milestone:** B · **Gate:** G-standard · **Depends on:** 001, 003

## Goal

Cover the two lowest-coverage plugins: `apikeys` (0%) and `reports` (24.4%).

## Context

`internal/plugins/apikeys` has no test files at all; its handlers
(`handleList`, `handleCreate`, `handleRotate`, `handleRevoke`, `cleanPermissions`,
`record`) and `permissionDefs` are unverified. `internal/plugins/reports` has
only two error-path tests (`TestHandleCreateRejectsMissingFields`,
`TestHandleCloseNotFound`); `handleMine`, `handleList`, the create happy path,
rate limiting and `permissionDefs` are untested.

## Requirements

- `apikeys`: list, create (secret returned once; stored hashed), rotate, revoke,
  `cleanPermissions`, `keyDTO`; assert scopes/permissions and that the secret is
  never returned in list/rotate. Cover `Register` and `permissionDefs`.
- `reports`: create happy path, `handleMine`, `handleList`, `handleClose`
  success, rate-limit behavior, `Register` and `permissionDefs`.
- Use `internal/testsupport` for request/decode/auth; keep per-package fakes.

## Tests

- Table-driven handler tests covering success, validation (`422`), not-found
  (`404`) and authorization.
- A regression test that a revoke/list response never contains the plaintext
  key.

## Acceptance criteria

- `apikeys` and `reports` each reach ≥ 60%.
- `task check` green.

## Out of scope

- The apikeys/reports repository wrappers (C9, integration).
