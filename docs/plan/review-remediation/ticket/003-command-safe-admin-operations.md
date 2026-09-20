# 003 — Command-safe admin and moderation operations

**Phase:** 1 · **Gate:** G-standard, contributes to G1-security · **Depends on:** 002

## Goal

Apply the command-safety helpers introduced in ticket 002 to every
administrative command, closing the second command-injection sink.

## Findings addressed

- Critical: `internal/plugins/azerothadmin/plugin.go` builds `.ban account`,
  `.unban account` and `.account set gmlevel` from `Username` (path), `Reason`,
  `Duration` and `Realm` with no validation.
- The same package already validates player/character names and sanitizes text
  in `moderation.go`, so the fix is to reuse that policy consistently.

## Context

Handlers read `r.PathValue("username")` directly (`http.go`) and JSON `Reason`,
`Duration`, `Realm`; the command builders only check for an empty username.
`Duration` defaults to `1d`, `Realm` defaults to `-1`.

## Atomic change

Validate every dynamic value in `azerothadmin`'s command builders using the
helpers from ticket 002. No new abstractions beyond what 002 introduced.

## Requirements

- `banAccount`: username via `SafeIdentifier`; `Duration` via a new
  `SafeBanDuration` (`^\d+[smhdw]$`, with a documented maximum) or the explicit
  empty default; `Reason` via `Quote`.
- `unbanAccount`: username via `SafeIdentifier`.
- `setGMLevel`: username via `SafeIdentifier`; `Level` bounded to the valid GM
  range (0–4 or the AzerothCore maximum); `Realm` validated as an integer
  (`-1` = all realms) or a known realm, never free text.
- Reuse the existing character-name validator and text sanitizer in
  `moderation.go` rather than duplicating them.
- Handlers must map validation failures to `422` with a stable error code.

## Tests

- Parameterized tests for each builder with injection payloads in every field
  (`Duration`: `1d; .server shutdown`, `../../`, empty; `Realm`: `all', ...`).
- Handler tests asserting `422` and no executor call for invalid input.
- Regression test asserting a valid ban still produces the exact expected
  command string.

## Acceptance criteria (gate G-standard, G1-security)

- `task check` green; `task test:race` green.
- No `fmt.Sprintf` command builder in `azerothadmin` uses an unvalidated value.
- `grep` for command builders shows every interpolation guarded by a helper.

## Rollback

Revert. As with 002, revert is only safe if the affected permissions are removed.

## Out of scope

- `azerothaccount` (ticket 002).
- Audit coverage for these actions (ticket 025).
