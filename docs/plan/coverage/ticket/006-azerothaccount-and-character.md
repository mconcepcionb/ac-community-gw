# azerothaccount and azerothcharacter

**Milestone:** B · **Gate:** G-standard · **Depends on:** 001, 003

## Goal

Raise `azerothaccount` (36.4%) and `azerothcharacter` (36.1%) to ≥ 60%,
covering the claim flow, self-service reads, leaderboards and visibility.

## Context

- `azerothaccount`: `claim.go` (start/verify, code generation and hashing,
  attempts, expiry, staff claim list) is entirely untested; also `audit.go`,
  `me.go`, `permissions.go` and `Register` (`internal/plugins/azerothaccount/`).
- `azerothcharacter`: `leaderboard.go` (`buildLeaderboard`, `boardWindow`,
  public opt-in and cache), `visibility.go`, `me.go`, `notice.go`,
  `directory.go`, `permissions.go`, `Register`
  (`internal/plugins/azerothcharacter/`). Existing tests cover listing, detail,
  mail and delivery.

## Requirements

- `azerothaccount`: claim start (code issued), verify (success, wrong code,
  expired, max-attempt lockout), staff claim listing; `me` self-service read;
  audit recording.
- `azerothcharacter`: leaderboard build and windowing, public board honoring the
  per-character opt-in, cache behavior; visibility get/set authorization;
  directory reads; `me`; notice.
- Cover both `Register`s and `permissionDefs`.

## Tests

- A claim test that a reused/guessed code is rejected and attempts are counted.
- A public-leaderboard test proving non-opted-in characters are absent.

## Acceptance criteria

- Both plugins reach ≥ 60%.
- Ownership/authorization paths have negative tests.
- `task check` green.

## Out of scope

- The repository wrappers (C9, integration).
- AzerothCore SOAP transport (already 80.6%).
