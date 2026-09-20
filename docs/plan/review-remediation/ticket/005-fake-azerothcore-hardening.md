# 005 — Fake AzerothCore hardening

**Phase:** 1 · **Gate:** G-standard, contributes to G1-security · **Depends on:** 001

## Goal

Make the development double safe by default: authenticate every endpoint, never
expose or log credentials, and keep its in-memory state and database mirror
consistent.

## Findings addressed

- Critical: `GET /state`, `/commands`, `/commands/stream`, `POST /reset` are not
  behind Basic Auth; `/state` serializes `Account.Password` in cleartext; CORS is
  `Access-Control-Allow-Origin: *`.
- High: command text containing generated passwords is logged and journaled.
- Medium: `POST /reset` resets memory but not the MySQL mirror, desynchronizing
  the system under test.
- Low: default credentials `acgw/acgw`; empty `-user` disables auth silently;
  `account set password` skips the max-length check; unknown ban duration units
  silently become permanent; non-200 upstream body snippets are embedded in
  errors/logs; the server mutex is held across MySQL network calls; the HTTP
  server lacks read/idle timeouts.

## Context

`internal/fake/azerothcore/*` and `cmd/fakeazerothcore/main.go`. This is a local
development double, but it is run from the same compose stack and its defaults
make accidental exposure dangerous.

## Atomic change

Wrap the fake server's handlers in an auth middleware, redact credentials from
all outputs, and correct the double's state/locking behaviour. Keep the SOAP
wire contract and journal format unchanged except for credential redaction.

## Requirements

- Register `/healthz` unauthenticated; every other route (including `/reset`,
  `/state`, `/commands`, `/commands/stream`, dashboard and `/`) goes through an
  auth check that reuses the constant-time comparison already present.
- When no credentials are configured, refuse to start unless the bind address is
  loopback (or an explicit `-insecure`/`-no-auth` flag is passed); never silently
  disable auth.
- Remove `Password` from the JSON snapshot (`json:"-"`) and redact password
  arguments in the journal and logs (keep the command verb and non-secret args).
- Restrict CORS to a configurable allowed origin (default loopback dev origins),
  not `*`.
- `Reset` must resynchronize the MySQL mirror (re-apply the seed and remove
  accounts created since startup) so the fake and the databases agree.
- Apply the same maximum-password check in `account set password` as in
  `account create`.
- Ban durations with unknown units must be rejected, never treated as permanent.
- Do not embed upstream response bodies in errors unless a debug flag is set.
- Release the server mutex before MySQL mirror writes.
- Add `ReadTimeout` and `IdleTimeout` to the fake HTTP server.

## Tests

- Auth test: unauthenticated `/state`, `/commands`, `/reset` return `401`;
  `/healthz` returns `200`.
- Snapshot test: no `password` key and no password substring anywhere in the JSON.
- Journal/log test: a created account's password never appears in the journal.
- Reset test: an account created after startup disappears from the mirror after
  `POST /reset`.
- Duration test: unknown units are rejected.
- Concurrency test: a slow mirror does not block `/state`.
- Update `docs/runbooks/fake-azerothcore.md` for the new auth/bind behaviour.

## Acceptance criteria (gate G-standard, G1-security)

- `task check` green; `go test -race ./...` green.
- No endpoint except `/healthz` is reachable without credentials.
- Searching the fake's outputs for a known password yields no match.

## Rollback

Revert. Reverting reintroduces the credential exposure; prefer fixing forward.

## Out of scope

- Changing the SOAP response format.
- The real SOAP adapter (ticket 019).
