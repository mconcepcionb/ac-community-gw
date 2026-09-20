# 002 — Command-safe account operations

**Phase:** 1 · **Gate:** G-standard, contributes to G1-security · **Depends on:** 001

## Goal

Prevent console/command injection through account-management operations by
introducing a single command-safety helper and using it on every value before a
command string is assembled.

## Findings addressed

- Critical: `internal/plugins/azerothaccount/plugin.go` interpolates `Username`,
  `Password` and `Email` from the request (path and JSON) into
  `.account create`, `.account set password` and `.account set email` without
  validation.
- Related: the SOAP executor does not reject commands containing control
  characters, so a single missed call site is exploitable.

## Context

`docs/security.md` and `docs/azerothcore-integration.md` promise that no generic
command endpoint exists and that values embedded in command strings are
sanitized. `email` is currently not validated at all, and username/password only
check for emptiness. `mail.go` and `moderation.go` already sanitize, but with
ad-hoc helpers.

## Atomic change

Add `internal/core/azerothcore/safe.go` with validated constructors for the
values used in console commands, and use them in the `azerothaccount` plugin.
Add a defense-in-depth rejection of CR/LF in the command transport.

## Requirements

- New `internal/core/azerothcore/safe.go` (core infrastructure, no domain
  knowledge):
  - `SafeIdentifier(value string) (string, error)` — `^[A-Za-z0-9_]{1,32}$`.
  - `SafeAccountPassword(value string) (string, error)` — reject control chars,
    whitespace, quotes and command metacharacters; enforce the AzerothCore
    maximum length; allow the printable password set.
  - `SafeEmail(value string) (string, error)` — a conservative email grammar; no
    whitespace or control characters.
  - `Quote(value string) (string, error)` — collapses CR/LF/quotes and caps
    length, for free-text command arguments.
- `createAccount`: validate username, password and email; if email is optional,
  use a safe placeholder or omit the argument rather than inject an empty token.
- `changePassword`: validate username and password.
- `setEmail`: validate username and email.
- The `azerothsoap` client (or the core executor decorator) rejects any command
  containing `\r` or `\n` with a typed error before transmission.
- Error messages returned by handlers must not echo the rejected value if it
  could contain the password.

## Tests

- Unit tests for `SafeIdentifier`, `SafeAccountPassword`, `SafeEmail`, `Quote`
  covering valid values and injection payloads (`foo bar`, `foo\n.server shutdown`,
  `"; drop`, empty, over-length, unicode).
- Handler/command tests for `.account create`/`set password`/`set email` with
  injection payloads asserting the executor is never called (or is called with a
  rejected error).
- A transport test asserting a CR/LF command is refused.

## Acceptance criteria (gate G-standard, G1-security)

- `task check` green; `go test -race ./...` green.
- A payload containing a newline or space can no longer reach the executor.
- Documentation for the helper is added to `docs/azerothcore-integration.md`.

## Rollback

Revert the ticket; the previously accepted-but-unsafe commands return. Revert is
only acceptable if the endpoint permissions are also revoked.

## Out of scope

- Admin/moderation commands (ticket 003).
- Rewriting the SOAP adapter (ticket 019).
