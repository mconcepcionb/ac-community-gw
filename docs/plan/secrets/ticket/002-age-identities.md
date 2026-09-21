# Age identities and creation rules

**Milestone:** A · **Gate:** G-standard · **Depends on:** 001

## Goal

Create the dedicated `ac-community-gw` age identities and commit the public
recipients in `.sops.yaml`.

## Context

The model is per-path `creation_rules`; each rule lists the recipients allowed
to decrypt that file. Identities must be dedicated to this project, not the
homelab's.

## Requirements

- Generate, with `age-keygen`:
  - a dedicated **admin** identity (workstation/administration);
  - a **dev-host** identity;
  - a placeholder **deploy-host** identity (finalized in S7).
- Store each private key outside the repository (documented location; never
  tracked, never in CI).
- Add `.sops.yaml` with `creation_rules` for
  `^secrets[/\\]development\.sops\.env$` (admin + dev host) and
  `^secrets[/\\]production\.sops\.env$` (admin + deploy host).
- Document the ceremony (generation, storage, recovery) in
  `docs/development.md` or a runbook.
- Confirm `sops` and `age` are available on the supported platforms; document
  installation for Windows and Linux/macOS.

## Tests

- `sops --version` and `age-keygen --version` succeed on the documented path.
- A guard (implemented in S3/S5) fails if an `AGE-SECRET-KEY-` string or an
  unencrypted `secrets/*` file is tracked.

## Acceptance criteria

- `.sops.yaml` present with the documented public recipients only.
- No private key is tracked; the recovery procedure is documented.

## Out of scope

- Encrypting the actual values (S3); production recipient finalization (S7).
