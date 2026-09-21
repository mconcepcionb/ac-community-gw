# Leak guard in CI

**Milestone:** C · **Gate:** G-standard · **Depends on:** 003

## Goal

Make a plaintext-secret leak fail CI, without giving CI the ability to decrypt.

## Context

CI references only `secrets.GITHUB_TOKEN` (`ci.yml:24`) and has no secret
manager. By decision, CI never decrypts; the integration job (coverage plan C9)
uses ephemeral credentials. The guard is therefore a static check, not a
decrypt-and-scan.

## Requirements

- Add `task secrets:check` to CI (a lightweight step in the existing Go job, or
  a small dedicated job).
- The check fails when:
  - a tracked file under `secrets/` is neither `*.env.example` nor
    `*.sops.env`;
  - a tracked file contains an `AGE-SECRET-KEY-` header;
  - a `*.sops.env` lacks SOPS metadata (best-effort plaintext detection).
- Ensure CI references no application secret and no age identity.
- Document the check as a required branch-protection check.

## Tests

- CI passes on the clean tree.
- A branch that commits `secrets/development.env` (or a private key) fails the
  check.

## Acceptance criteria

- The check is required and green on `main`.
- CI has no application secret and no decryption capability.

## Out of scope

- Decryption anywhere in CI (explicitly rejected).
- Rotation (S6).
