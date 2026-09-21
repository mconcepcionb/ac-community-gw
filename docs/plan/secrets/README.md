# Secret management with SOPS + age

## Goal

Remove plaintext secrets from the developer and operator workflow: secrets are
committed **encrypted** (SOPS + age), decrypted only on the machine that needs
them, with the private identities held outside the repository. Rotate the
credentials that were exposed in the local `.env` as part of the rollout.

## Status

| Area | State |
| --- | --- |
| Threat model and recipient design | planned |
| `.sops.yaml` and key ceremony | planned |
| Encrypt the development secret set | planned |
| Local/compose materialization flow | planned |
| CI decryption and plaintext leak check | planned |
| Rotate exposed credentials and runbooks | planned |
| Deploy-host recipient and ADR | planned |

## Context

Today secrets live in a plaintext, gitignored `.env` that Task loads through
`dotenv`, and Compose requires them as interpolation variables with no defaults
(`compose.yaml`). The [review-remediation plan](../review-remediation/README.md)
recorded, as an operational follow-up, that the Discord client secret and the
SOAP password were present in the local `.env` and must be rotated, and that
Compose must be given `POSTGRES_PASSWORD` / `MARIADB_*` explicitly.

There is no secret manager. The sibling `homelab-config` repository already runs
a proven **SOPS + age** model that this plan reuses:

- `.sops.yaml` defines `creation_rules` keyed by path, each listing age
  recipients (an admin workstation identity plus the runtime node identity).
- Only encrypted `*.sops.env` files are committed; `.gitignore` allows
  `*.sops.env` and `*.env.example` and excludes everything else under
  `secrets/`.
- Private age identities live outside the repo (`/etc/homelab/age/keys.txt` on
  the nodes, and the workstation's own copy); `sops <file>` edits in place and
  `scripts/materialize-secrets.sh` decrypts a node's secrets into a
  permission-restricted runtime directory.
- Decryption is `sops decrypt --input-type dotenv --output-type dotenv`.

This plan adopts the same model, adapted to a single-application repository and
to the Compose/Task/CI toolchain used here. The decision is recorded as an ADR
(see ticket S1).

## Scope

- A `.sops.yaml` with `creation_rules` for this repo's encrypted secret files.
- An age **recipient/key ceremony**: an admin workstation identity and a runtime
  identity, private keys never committed.
- Encrypted secret files committed under `secrets/` (`*.sops.env`) plus
  `*.env.example` templates in plaintext.
- Scripts and Task tasks to create, edit, decrypt and materialize secrets
  locally and for Compose.
- CI: decrypt with an age identity supplied as a GitHub Actions secret, and a
  job that fails if a plaintext secret is committed.
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

## Principles

1. **No plaintext secret in the repo.** Only `*.sops.env` (encrypted) and
   `*.env.example` (templates, no real values) are committed.
2. **Identities never travel with the ciphertext.** Private age keys live
   outside the repo and outside CI logs; CI gets one via a masked secret.
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
scripts/secrets/
  check-no-plaintext.sh          fail if a non-.sops/.env.example secret is tracked
  decrypt-env.sh                 decrypt one file to stdout / a target path
  materialize.sh                 decrypt all for an environment to a runtime dir
```

`.gitignore` gains the repository equivalent of the homelab rule: ignore
everything under `secrets/` except `*.sops.env` and `*.env.example`; keep the
existing `.env` / `.env.*` ignores.

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
      <production-host-recipient>
```

Recipients are per trust domain: developers + the dev host for
`development.sops.env`; the admin + the production host for
`production.sops.env`. CI does not get a dedicated recipient by default; when it
needs real secrets it uses the development identity stored as a masked GitHub
Actions secret (ticket S4) rather than a new recipient.

### Key ceremony

- Generate the admin identity once (`age-keygen`) and store the private key
  outside the repo (operator password manager / secure disk); commit only its
  public recipient in `.sops.yaml`.
- Generate one identity per runtime host (dev workstation/host, production
  host) the same way; the private key is deployed to that host and referenced by
  `SOPS_AGE_KEY_FILE`.
- Record recipients in `.sops.yaml`; never record private keys, not even
  encrypted.

### Editing and materializing

- Edit: `sops secrets/development.sops.env` (decrypts to a temp file, re-encrypts
  on save). `sops filestatus` confirms encryption.
- Local dev: `task secrets:decrypt ENV=development` writes `.env` (gitignored)
  from the encrypted file, preserving the current Task `dotenv` workflow.
- Compose: `task secrets:materialize ENV=development` writes the decrypted files
  into a runtime directory consumed by Compose, mirroring
  `materialize-secrets.sh`, with restrictive permissions.
- CI: `task secrets:decrypt` in a job that needs secrets, with the identity
  injected from a masked secret; the plaintext file is never uploaded as an
  artifact.

### Leak guard

`scripts/secrets/check-no-plaintext.sh` runs in CI and as a `task secrets:check`:
it fails if any tracked file matching the secret set is not a `*.sops.env` or
`*.env.example`, and (best-effort) if an encrypted file appears to contain
plaintext (no `sops` metadata). This replaces the current honour-system `.env`
ignore.

## Backend gaps

| Gap | Needed by |
| --- | --- |
| `.gitignore` rule allowing `*.sops.env` | S2 |
| Scripts + `task secrets:*` | S2, S3 |
| Compose/CI consumption of decrypted secrets | S3, S4 |
| Rotation runbooks updated for SOPS | S5 |
| ADR recording the decision | S1 |

## Milestones

| Milestone | Tickets |
| --- | --- |
| A - Design and keys | S1, S2 |
| B - Dev and compose flow | S3 |
| C - CI and leak guard | S4 |
| D - Rotation and production | S5, S6 |

### Dependency graph

```
S1 -> S2 -> S3
S2 -> S4
S2 -> S5
S3, S4 -> S6
```

S1 fixes the trust model before any file is encrypted. S2 introduces the files
and scripts; S3 and S4 consume them locally and in CI. Rotation (S5) can start
once the encrypted set exists, so the rotated values land encrypted from the
start.

## Risks

| Risk | Mitigation |
| --- | --- |
| An age private key is committed | `.gitignore` + `check-no-plaintext.sh`; key ceremony stores keys outside the repo; CI scans for age private key headers |
| CI cannot decrypt (missing identity) | The identity is a masked Actions secret; jobs that need it declare it, others stay secret-free |
| Disabling `.env` breaks the Task workflow | S3 keeps `task secrets:decrypt` producing `.env`, so `dotenv` is unchanged |
| Recipients are too broad | Per-path creation rules and per-domain recipients; only the affected domain can decrypt |
| Encrypting gives false safety and rotation is skipped | S5 is a required ticket and rotates the two known-exposed credentials |
| Windows developers lack `sops`/`age` | Scripts ship as both `.ps1` and `.sh`, and `docs/development.md` documents the install |

## Tickets

### Milestone A - Design and keys

- **S1 - ADR and threat model.** Write ADR `0015-sops-age-secrets.md`: why
  SOPS + age, the trust domains, the recipient set, what is encrypted and what
  stays as an Actions secret. *Gate:* ADR accepted; recipients agreed.

- **S2 - `.sops.yaml`, files and scripts.** Add `.sops.yaml`, the `secrets/`
  layout, the plaintext templates, the `.gitignore` rules and
  `scripts/secrets/` (encrypt/edit guidance, decrypt, materialize, leak check)
  in both `sh` and `ps1`. Encrypt the existing development secret set (with the
  exposed values still in place; rotation is S5). *Gate:* `task secrets:check`
  green; `sops filestatus` shows the committed files encrypted; `docker compose
  config` still works from decrypted materialized values.

### Milestone B - Dev and compose flow

- **S3 - Local and Compose materialization.** Add `task secrets:decrypt` and
  `task secrets:materialize`; document the flow; make `task dev`, `task run` and
  `docker compose up` consume the decrypted output without changing the env-var
  contract. *Gate:* a clean checkout plus the operator's age key brings up dev
  and Compose; no plaintext secret on disk beyond the gitignored decrypted file.

### Milestone C - CI and leak guard

- **S4 - CI decryption and leak check.** Add a CI job that decrypts when a
  required identity secret is present and runs the checks that need real
  secrets; add the leak guard as a required check. *Gate:* CI green without
  exposing values; leak guard fails on a deliberately committed plaintext
  secret.

### Milestone D - Rotation and production

- **S5 - Rotate exposed credentials.** Rotate the Discord client secret and the
  SOAP password, update them in the encrypted set, and update
  `docs/runbooks/rotate-discord-secret.md` and
  `docs/runbooks/rotate-azeroth-soap-credentials.md` to the SOPS flow. *Gate:*
  old values rejected; runbooks reproduce the rotation.

- **S6 - Production recipient and documentation.** Add the production host
  recipient and `production.sops.env`, wire the deploy path, and point
  `docs/security.md`, `docs/development.md` and `docs/README.md` at the flow.
  *Gate:* a production-like host decrypts only its own file; docs match the
  shipped behaviour.
