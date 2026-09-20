# 016 — Open redirect hardening (server)

**Phase:** 3 · **Gate:** G-standard, contributes to G3-hardening · **Depends on:** 001

## Goal

Close the `return_to` open redirect on the server side by rejecting backslashes
and validating the target as a same-origin path.

## Findings addressed

- Medium: `sanitizeReturnTo` rejects `//` and CR/LF but allows `/\evil.test`;
  browsers normalize `\` to `/`, turning it into protocol-relative `//evil.test`.

## Context

`internal/plugins/identitydiscord/oauth.go`. The SPA mirrors the same flawed
helper (fixed in ticket 034). The callback redirects to the sanitized value.

## Atomic change

Replace the string-prefix check with a parse-and-validate helper that accepts
only same-origin absolute paths.

## Requirements

- Reject any value containing `\`.
- Parse with `net/url`; require an empty host and scheme, a path starting with
  `/`, and no `//` after normalization.
- Reject control characters, whitespace and encoded separators (`%5c`, `%2f`)
  that normalize to a dangerous form.
- Keep the empty-value behaviour (return `""`).
- Document the policy next to the function.

## Tests

- Table tests: valid (`/characters`, `/store/wallet`) and invalid
  (`https://evil.test`, `//evil.test`, `/\evil.test`, `/\\evil.test`,
  `/%5Cevil.test`, `javascript:alert(1)`, `\r\n`, empty).
- Callback test: an unsafe `return_to` falls back to the configured default, not
  the unsafe value.

## Acceptance criteria (gate G-standard, G3-hardening)

- `task check` green.
- No value containing a backslash can influence the `Location` header.

## Rollback

Revert; the backslash bypass returns.

## Out of scope

- The SPA helper (ticket 034).
