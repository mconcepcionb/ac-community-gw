# 025 — Audit coverage for account and admin actions

**Phase:** 4 · **Gate:** G-standard, contributes to G4-data · **Depends on:** 002, 003

## Goal

Record an audit entry for every privileged command-issuing endpoint.

## Findings addressed

- Medium: account creation, password change, email change, GM-level change,
  ban/unban, kick, mute/unmute and announce execute privileged SOAP commands with
  no `audit.Record` call. Other paths (links, mail, purchase) already record
  correctly.

## Context

`internal/plugins/azerothaccount/http.go` and `internal/plugins/azerothadmin/*`.
`audit.Recorder` and the `audit.Entry` contract already exist;
`docs/security.md` defines the fields.

## Atomic change

Add audit recording to the missing handlers using the established pattern
(actor, action, permission, target type/id, result, request id). Never record
passwords, tokens or session ids.

## Requirements

- Record on both success and failure for:
  - `account.create`, `account.change-password`, `account.set-email`
  - `account.ban`, `account.unban`, `account.set-gmlevel`
  - `azeroth.players.kick`, `players.mute`, `players.unmute`
  - `azeroth.characters.ban`, `characters.unban`
  - `azeroth.announce`
- `TargetType`/`TargetID` identify the account/character/player; the announce
  target is the realm/channel, if any.
- `Result` reflects success/failure; `RequestID` comes from the context.
- Never place the password or email body in `Metadata`/the entry; use identifiers.
- If ticket 015 fixed the audit recorder to include `ActorDiscordID`/`Metadata`,
  use those fields consistently.

## Tests

- Table test per handler: success records one entry with the right action/result;
  failure records a failure entry.
- Test: no recorded entry contains the password string used in the request.
- Test: an unauthenticated request never reaches the audit path.

## Acceptance criteria (gate G-standard, G4-data)

- `task check` green.
- Every command-issuing handler records exactly one audit entry per outcome.

## Rollback

Revert; the missing audit coverage returns. Low runtime risk.

## Out of scope

- Audit persistence backend (currently the log recorder).
- Audit retention (ticket 021).
