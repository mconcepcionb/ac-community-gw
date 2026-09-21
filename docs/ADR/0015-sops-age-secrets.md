# ADR 0015: SOPS and age for secret management

## Status

Accepted. The key ceremony (private identities) is an operational step recorded
in the [secrets plan](../plan/secrets/README.md); no private key is ever stored
in this repository.

## Context

Configuration is read from environment variables. Locally that means a plaintext
`.env` loaded by Task (`Taskfile.yml`) — gitignored, but present on disk — and
Compose requires the database passwords as interpolation variables. The
[code-review remediation](../plan/review-remediation/ticket/029-compose-secrets.md)
recorded that a live-looking Discord client secret and the AzerothCore SOAP
password were present in that local `.env`, and that they must be rotated.

There is no secret manager, no encryption and no `sops`/`age`/vault reference in
code or configuration. Secrets therefore exist only in plaintext files that a
mistake (a stray `git add`, a build context, a backup, a log) can leak. The
gateway must also be deployable to a host that needs the same secrets without a
human present.

The sibling `homelab-config` repository already runs a proven model: encrypted
`*.sops.env` files committed, `age` recipients declared per path in `.sops.yaml`,
private identities kept outside the repository, and a materialize script that
decrypts to a restricted runtime directory.

## Decision

Adopt **SOPS + age** for secrets in this repository, isolated from the homelab
trust domain.

- **Encrypted files are committed; plaintext is not.** Secret sets live at
  `secrets/<environment>.sops.env` (encrypted) with a matching
  `secrets/<environment>.env.example` (names only, no values). Only these two
  patterns are tracked under `secrets/`.
- **One creation rule per trust domain.** `.sops.yaml` lists, per path, the age
  recipients allowed to decrypt it. The `development` set is readable by the
  admin workstation and the development host; the `production` set by the admin
  workstation and the deploy host.
- **Dedicated identities.** The project has its own age identities (admin,
  dev host, deploy host); it does **not** reuse the homelab keys. Private keys
  live outside the repository (operator store and the runtime hosts) and are
  never committed, not even encrypted.
- **CI never decrypts.** CI jobs are secret-free by design; the integration job
  uses ephemeral service-container credentials. A static leak guard fails CI if
  a plaintext secret or an `AGE-SECRET-KEY-` value is tracked.
- **Editing is one command.** A human edits a secret with `sops <file>`; a
  machine decrypts with a script/task (`secrets:decrypt`, `secrets:materialize`).
- **Rotation is part of adoption.** The exposed Discord client secret and SOAP
  password are rotated, and the rotation runbooks are rewritten to edit the
  encrypted file.

SOPS protects the **committed artifact**, not runtime interpolation: Compose and
the process still receive plaintext environment values at run time. This is
accepted; the aim is to remove plaintext secrets from version control and from
the developer's working tree, not to build a runtime secret broker.

## Consequences

- No plaintext secret can be committed without failing the leak guard.
- A new developer or a fresh host needs an age identity granted by an admin
  recipient; access is revoked by removing that recipient and re-encrypting.
- `.env` remains the runtime mechanism (produced by decrypting the encrypted
  set), so the existing `dotenv`, Compose and startup-validation behaviour is
  unchanged.
- Windows and POSIX are both supported: the helper scripts ship in `sh` and
  `ps1`, and the tooling (`sops`, `age`) is available on both.
- Losing the admin identity makes the ciphertext unrecoverable unless another
  recipient still holds it; the admin key must be backed up out of band.
- The repository no longer needs any GitHub Actions secret for application
  configuration.
- A hosted secret manager (Vault, cloud KMS) is explicitly not adopted; it can
  be revisited if a runtime broker is ever required.

## See also

- [../plan/secrets/README.md](../plan/secrets/README.md)
- [../security.md](../security.md)
- [../development.md](../development.md)
- [../plan/review-remediation/ticket/029-compose-secrets.md](../plan/review-remediation/ticket/029-compose-secrets.md)
- `homelab-config` (sibling repository): the `.sops.yaml` and
  `scripts/materialize-secrets.sh` model this decision reuses in shape.
