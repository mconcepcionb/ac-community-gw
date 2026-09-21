# Secret management with SOPS + age

## Goal

Remove plaintext secrets from the developer and operator workflow: secrets are
committed **encrypted** (SOPS + age), decrypted only on the machine that needs
them, with the private identities held outside the repository. Rotate the
credentials that were exposed in the local `.env` as part of the rollout.

## Status

| Area | State |
| --- | --- |
| ADR, threat model and recipient design | implemented (S1) |
| Age identity ceremony and `.sops.yaml` | implemented (S2) |
| Encrypted secret set, scripts and `.gitignore` | planned (S3) |
| Local/compose materialization flow | planned (S4) |
| Plaintext leak guard in CI | planned (S5) |
| Rotate exposed credentials and runbooks | planned (S6) |
| Deploy-host recipient and documentation | planned (S7) |

## Decisions

- **A new dedicated `ac-community-gw` age identity** is used (not the homelab
  recipient): one admin identity plus a deploy-host identity, isolated from the
  `homelab-config` trust domain. The homelab model is reused, not its keys.
- **CI never decrypts.** CI stays secret-free; the integration job (see the
  coverage plan, C9) uses ephemeral service-container credentials. SOPS is for
  the developer and operator/deploy flow.
- **ADR 0015** records the decision (the next free ADR number).
- **Committed artifact rule:** only `secrets/*.sops.env` (encrypted) and
  `*.env.example` (templates, no real values) may be tracked; everything else
  under `secrets/` is ignored.

## Context

Today secrets live in a plaintext, gitignored `.env` that Task loads through
`dotenv` (`.gitignore:19`, `Taskfile.yml:3`), and Compose requires DB passwords
as interpolation variables with no defaults while treating the Discord secret
and SOAP password as optional empty defaults (`compose.yaml:52,59`). The
[review-remediation plan](../review-remediation/README.md) recorded, as an
operational follow-up, that the Discord client secret and the SOAP password are
present in the local `.env` and must be rotated, and that Compose must be given
`POSTGRES_PASSWORD` / `MARIADB_*` explicitly.

There is no secret manager, no encryption and no `sops`/`age`/vault reference in
code or config; the only mentions are this plan and the future-tense notes in
[security.md](../../security.md:56) and [development.md](../../development.md:24).
`.env` is correctly untracked and excluded from the Docker build context
(`.dockerignore:5-7`), but a live-looking `.env` exists on disk.

The sibling `homelab-config` repository already runs a proven **SOPS + age**
model this plan reuses in shape:

- `.sops.yaml` defines `creation_rules` keyed by path, each listing age
  recipients (an admin workstation identity plus the runtime node identity).
- Only encrypted `*.sops.env` files are committed; `.gitignore` allows
  `*.sops.env` and `*.env.example` and excludes everything else under the secret
  paths.
- Private age identities live outside the repo; `sops <file>` edits in place and
  `scripts/materialize-secrets.sh` decrypts a node's secrets into a
  permission-restricted runtime directory (`sops decrypt --input-type dotenv
  --output-type dotenv`).

## Scope

- A `.sops.yaml` with `creation_rules` for this repo's encrypted secret files.
- An age **recipient/key ceremony**: a dedicated admin identity and a deploy-host
  identity; private keys never committed.
- Encrypted secret files committed under `secrets/` (`*.sops.env`) plus
  `*.env.example` templates in plaintext.
- Scripts and Task tasks to create, edit, decrypt and materialize secrets
  locally and for Compose.
- A CI **plaintext leak guard**; CI never decrypts.
- Rotation of the exposed Discord client secret and SOAP password, updating the
  existing rotation runbooks.
- Docs: `docs/security.md`, `docs/development.md` and the runbooks describe the
  SOPS flow as the way secrets are handled.

## Out of scope

- A hosted secret manager (Vault, cloud KMS, 1Password). SOPS + age is the
  chosen mechanism; a hosted manager can be reconsidered later.
- Runtime secret injection into the gateway beyond an env file: the gateway
  keeps reading configuration from environment variables.
- Application-level envelope encryption or database field encryption.
- Decrypting secrets in CI.

## Principles

1. **No plaintext secret in the repo.** Only `*.sops.env` (encrypted) and
   `*.env.example` (templates, no real values) are committed.
2. **Identities never travel with the ciphertext.** Private age keys live
   outside the repo and outside CI; CI gets none.
3. **One creation rule per trust domain.** The recipients are the set of
   machines/people allowed to decrypt that file, and nothing broader.
4. **Rotate, don't just encrypt.** An exposed value stays exposed after
   encryption; the rollout includes rotating the known-leaked credentials.
5. **The workflow stays one command.** Editing a secret is `sops <file>`;
   bringing secrets to a machine is one script/task.

## Design

### Files and layout

```
.sops.yaml                       creation rules (age recipients per path)
secrets/
  development.sops.env           encrypted dev secret set (committed)
  development.env.example        template with variable names (committed)
  production.sops.env            encrypted prod secret set (committed)
  production.env.example         template (committed)
  development.env                decrypted output (gitignored)
scripts/secrets/
  materialize.sh / .ps1          decrypt all for an environment to a target dir
  decrypt-env.sh / .ps1          decrypt one file to stdout / a target path
  check-no-plaintext.sh / .ps1   fail if a non-.sops/.env.example secret is tracked
```

`.gitignore` gains the repository equivalent of the homelab rule:

```gitignore
secrets/**
!secrets/*.env.example
!secrets/*.sops.env
```

The existing `.env` / `.env.*` ignores stay. Note that `.env.*` does not match
`development.env`, which is why the explicit `secrets/**` rule is required.

### Creation rules

```yaml
creation_rules:
  - path_regex: '^secrets[/\\]development\.sops\.env$'
    age: >-
      <admin-workstation-recipient>,
      <dev-runtime-recipient>
  - path_regex: '^secrets[/\\]production\.sops\.env$'
    age: >-
      <admin-workstation-recipient>,
      <deploy-host-recipient>
```

Recipients are per trust domain: admin + dev host for `development.sops.env`;
admin + deploy host for `production.sops.env`. CI has no recipient.

### Key ceremony

- Generate the dedicated admin identity once (`age-keygen`) and store the
  private key outside the repo (operator password manager / secure disk); commit
  only its public recipient in `.sops.yaml`.
- Generate one identity per runtime host the same way; the private key is
  deployed to that host and referenced by `SOPS_AGE_KEY_FILE`.
- Record recipients in `.sops.yaml`; never record private keys, not even
  encrypted. Document the recovery procedure for the admin identity.

### Editing and materializing

- Edit: `sops secrets/development.sops.env` (decrypts to a temp file, re-encrypts
  on save). `sops filestatus` confirms encryption.
- Local dev: `task secrets:decrypt ENV=development` writes `.env` (gitignored)
  from the encrypted file, preserving the current Task `dotenv` workflow.
- Compose: `task secrets:materialize ENV=development` writes the decrypted file
  into the location Compose consumes, with restrictive permissions, mirroring
  `materialize-secrets.sh`.

### Leak guard

`scripts/secrets/check-no-plaintext.sh` (and `.ps1`) runs in CI and as
`task secrets:check`: it fails if any tracked file under `secrets/` is not a
`*.sops.env` or `*.env.example`, and (best-effort) if a `*.sops.env` lacks SOPS
metadata. This replaces the current honour-system `.env` ignore.

## Backend gaps

| Gap | Needed by |
| --- | --- |
| ADR 0015 recording the decision | S1 |
| Dedicated age identities and `.sops.yaml` | S2 |
| `.gitignore` rules allowing `*.sops.env` only | S3 |
| Scripts + `task secrets:*` | S3, S4 |
| Leak guard in CI | S5 |
| Rotation runbooks updated for SOPS | S6 |

## Milestones

| Milestone | Tickets |
| --- | --- |
| A - Design and keys | S1, S2 |
| B - Files and dev flow | S3, S4 |
| C - Leak guard | S5 |
| D - Rotation and deploy | S6, S7 |

### Dependency graph

```
S1 -> S2 -> S3 -> S4
S3 -> S5
S3 -> S6
S2, S3 -> S7
```

S1 fixes the trust model before any file is encrypted. S2 creates the keys and
the creation rules; S3 introduces the files and scripts; S4 and S5 consume them.
Rotation (S6) starts once the encrypted set exists, so the rotated values land
encrypted from the start.

## Risks

| Risk | Mitigation |
| --- | --- |
| An age private key is committed | `.gitignore` + `check-no-plaintext.sh`; keys stored outside the repo; CI scans for age private key headers |
| CI cannot decrypt (missing identity) | By decision CI has no identity and never decrypts; integration uses ephemeral credentials |
| Disabling `.env` breaks the Task workflow | S4 keeps `task secrets:decrypt` producing `.env`, so `dotenv` is unchanged |
| Recipients are too broad | Per-path creation rules and per-domain recipients; only the affected domain can decrypt |
| Encrypting gives false safety and rotation is skipped | S6 is a required ticket and rotates the two known-exposed credentials |
| Windows developers lack `sops`/`age` | Scripts ship as both `.ps1` and `.sh`, and `docs/development.md` documents the install |
| Compose still echoes interpolated values | SOPS protects the committed artifact, not runtime interpolation; S4 documents the boundary and keeps the `.env`-style flow explicit |

## Tickets

### Milestone A - Design and keys

1. [001-adr-and-threat-model.md](ticket/001-adr-and-threat-model.md) - ADR 0015, trust domains, recipients, CI secret-free
2. [002-age-identities.md](ticket/002-age-identities.md) - dedicated admin + host identities, `.sops.yaml`

### Milestone B - Files and dev flow

3. [003-encrypted-set-and-scripts.md](ticket/003-encrypted-set-and-scripts.md) - `secrets/`, templates, scripts, `.gitignore`
4. [004-dev-and-compose-flow.md](ticket/004-dev-and-compose-flow.md) - `task secrets:decrypt|materialize`, Compose alignment

### Milestone C - Leak guard

5. [005-leak-guard-ci.md](ticket/005-leak-guard-ci.md) - `task secrets:check` as a required CI check

### Milestone D - Rotation and deploy

6. [006-rotate-exposed-credentials.md](ticket/006-rotate-exposed-credentials.md) - rotate Discord + SOAP, update runbooks
7. [007-deploy-host-and-docs.md](ticket/007-deploy-host-and-docs.md) - deploy recipient, docs, production set
