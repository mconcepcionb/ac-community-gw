# 004 — Mail delivery object-level authorization

**Phase:** 1 · **Gate:** G-standard, contributes to G1-security · **Depends on:** 001

## Goal

Stop any holder of `azeroth.mail.send` from mailing items or money to characters
they do not own.

## Findings addressed

- High: `internal/plugins/azerothcharacter/mail.go` checks only that the
  recipient character exists; it never checks ownership. The plugin's own
  `delivery.Deliver` capability does enforce ownership (`delivery.go`), so the
  HTTP path is an authorization bypass (IDOR / economy exploit).

## Context

`handleSendMail` resolves `p.characters.FindCharacter(recipient)` and proceeds if
found. The caller's linked account is available through the `azeroth.account`
directory (`AccountDirectory.LinkedAccount`) and the character reader exposes
`AccountID`. `delivery.go` already returns `delivery.ErrNotOwner`.

## Atomic change

Route the ownership decision through the same rule the delivery capability uses:
fetch the recipient, resolve the caller's linked account, and require
`character.AccountID == callerAccountID`. Fail closed when the character reader
is unavailable.

## Requirements

- Inject the account directory (or the delivery service) into the mail handler.
- Resolve the caller's account from the authenticated principal; return
  `403`/`422 account_not_linked` when the caller has no linked account.
- Return `404 character_not_found` when the character does not exist, and
  `403 not_owner` when it exists but belongs to another account.
- If the character reader is nil, return `503` (fail closed) instead of skipping
  the check, matching `delivery.go`.
- Do not leak whether a character exists on another account beyond the `403`
  already implied by the ownership model; document the chosen disclosure policy.
- Keep the existing name-pattern validation and text sanitization.

## Tests

- Handler test: caller mails own character → 200 and executor called.
- Handler test: caller mails another account's character → 403, no executor call.
- Handler test: caller with no linked account → 422/403, no executor call.
- Handler test: character reader nil → 503, no executor call.
- Integration-style test using the fake AzerothCore seed to prove ownership
  resolution end to end.

## Acceptance criteria (gate G-standard, G1-security)

- `task check` green; `go test -race ./...` green.
- The HTTP mail path and `delivery.Deliver` share the same ownership rule (or the
  HTTP path delegates to `Deliver`).

## Rollback

Revert. This reopens the IDOR; revert only with the endpoint disabled.

## Out of scope

- Changing who may hold `azeroth.mail.send`.
- Store purchase delivery (already owned through `delivery.Deliver`).
