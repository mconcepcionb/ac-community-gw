# Rotate exposed credentials

**Milestone:** D · **Gate:** G-standard · **Depends on:** 003

## Goal

Rotate the Discord client secret and the AzerothCore SOAP password that were
present in the local `.env`, and update the rotation runbooks to the SOPS flow.

## Context

The [review-remediation plan](../../review-remediation/README.md:25-27) recorded
the exposure as an operational follow-up; `.env` on disk still holds a live
32-char Discord secret and the SOAP password. Encryption alone does not revoke
an exposed value.

## Requirements

- Rotate the Discord client secret in the Discord developer portal.
- Create/rotate the SOAP credentials in AzerothCore.
- Update `secrets/development.sops.env` (via `sops`) with the new values;
  materialize and restart.
- Rewrite `docs/runbooks/rotate-discord-secret.md` and
  `docs/runbooks/rotate-azeroth-soap-credentials.md` so the "update the value"
  step edits the encrypted file and materializes it.
- Revoke the old credentials after validation.

## Tests

- Old Discord secret rejected by the token endpoint; `GET /readyz` green with
  the new SOAP credentials.
- A read-only SOAP capability succeeds end to end.

## Acceptance criteria

- The exposed values no longer work.
- Runbooks reproduce the rotation using SOPS only.
- No plaintext secret is committed (`task secrets:check` green).

## Out of scope

- Rotating local development DB passwords (they are not shared secrets); note
  this in the ticket.
