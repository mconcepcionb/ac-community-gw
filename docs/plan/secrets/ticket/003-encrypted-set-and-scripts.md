# Encrypted set, scripts and gitignore

**Milestone:** B · **Gate:** G-standard · **Depends on:** 002

## Goal

Introduce the `secrets/` layout, the encrypted development secret set, the
helper scripts and the ignore rules that keep plaintext out of git.

## Context

`.gitignore` currently ignores `.env` and `.env.*` (allowing only
`.env.example`); it has no rule for a decrypted `secrets/development.env`. The
homelab `materialize-secrets.sh` is the shape to mirror.

## Requirements

- Add `secrets/development.env.example`: variable names only, no values (reuse
  the key list from `.env.example`).
- Encrypt the current development secret set into
  `secrets/development.sops.env` (with the still-exposed values; rotation is
  S6). Commit only the encrypted file.
- Add scripts (both `sh` and `ps1`):
  - `scripts/secrets/decrypt-env` — decrypt one file to stdout/a path;
  - `scripts/secrets/materialize` — decrypt to a runtime/target location with
    restrictive permissions;
  - `scripts/secrets/check-no-plaintext` — fail if a tracked `secrets/**` file is
    not `*.env.example`/`*.sops.env`, or if a tracked file contains an
    `AGE-SECRET-KEY-` header.
- Add `.gitignore` rules:
  ```gitignore
  secrets/**
  !secrets/*.env.example
  !secrets/*.sops.env
  ```
- Add `task secrets:encrypt|decrypt|materialize|check` wrappers.
- Add `secrets/` to `.dockerignore`.

## Tests

- `sops filestatus secrets/development.sops.env` reports encrypted.
- `task secrets:check` green on the clean tree; fails on a planted plaintext
  file and on a planted `AGE-SECRET-KEY-` line.
- `task secrets:decrypt ENV=development` reproduces the expected variables.

## Acceptance criteria

- Only `*.sops.env` and `*.env.example` are tracked under `secrets/`.
- The scripts run on Windows (`ps1`) and POSIX (`sh`).
- `docker build` context excludes `secrets/`.

## Out of scope

- Rotation (S6), CI wiring (S5), production set (S7).
