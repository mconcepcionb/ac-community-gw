# Secret management (SOPS + age)

## Purpose

How secrets are stored, edited and delivered for `ac-community-gw`. Secrets are
committed **encrypted** with [SOPS](https://getsops.io/) and
[age](https://age-encryption.org/); private identities never enter the
repository. See [ADR 0015](../ADR/0015-sops-age-secrets.md) for the decision.

## Identities

Three dedicated age identities, one per trust domain. Private keys live
**outside the repository**:

| Identity | Private key location (workstation) | Purpose |
| --- | --- | --- |
| admin | `~/.config/ac-community-gw/age/admin.key.txt` | administration and recovery; a recipient of every secret set |
| dev host | `~/.config/ac-community-gw/age/dev-host.key.txt` | decrypts `secrets/development.sops.env` at run time |
| deploy host | `~/.config/ac-community-gw/age/deploy-host.key.txt` | decrypts `secrets/production.sops.env` at run time |

Their **public recipients** are the only key material recorded in the
repository (`.sops.yaml`):

```
admin       age185rgzn80jkd52v5qkn83nvm5zura2fz40kzgvuxp93p5x0xe24ks74qj6g
dev host    age16xss2yhvxhf2y7k7gy05rtn8k7v74rjp9vnq8vg8ddflln45w3gsrxn9n4
deploy host age1056k70l74x3thhkvkv7f2csj7t2g5gw95c575qlr354mqnaal45qyydsh3
```

## Ceremony

Generate an identity (one per trust domain; do this once):

```sh
age-keygen -o ~/.config/ac-community-gw/age/admin.key.txt
```

`age-keygen` writes the private key to the file and prints the public key. Record
only the public recipient in `.sops.yaml`. On Windows the same command works
(`age-keygen -o %USERPROFILE%\.config\ac-community-gw\age\admin.key.txt`).

The deploy-host private key is generated for the project now and must be moved
to the deploy host (and removed from the workstation) when that host is stood
up. See the [secrets plan](../plan/secrets/README.md), ticket S7.

## Recovery

There is no key escrow. If a private identity is lost, a secret set is
recoverable only while another listed recipient still holds its key: decrypt
with that identity, add a fresh recipient to `.sops.yaml`, and re-encrypt with
`sops updatekeys`. Back up the admin identity out of band; losing every
recipient of a set makes it unrecoverable.

## Editing a secret

```sh
sops secrets/development.sops.env          # decrypts to a temp file, re-encrypts on save
sops filestatus secrets/development.sops.env  # confirms the file is encrypted
```

After changing recipients in `.sops.yaml`:

```sh
sops updatekeys secrets/development.sops.env
```

## Delivery

SOPS resolves the age identity from `SOPS_AGE_KEY_FILE` (or its default,
`~/.config/sops/age/keys.txt`). Each host points it at its own identity:

| Host | `SOPS_AGE_KEY_FILE` |
| --- | --- |
| admin workstation | `~/.config/ac-community-gw/age/admin.key.txt` |
| dev host | `~/.config/ac-community-gw/age/dev-host.key.txt` |
| deploy host | `~/.config/ac-community-gw/age/deploy-host.key.txt` |

Then:

```sh
task secrets:edit ENV=development        # edit the encrypted set in place
task secrets:decrypt ENV=development     # decrypt secrets/development.sops.env into .env
task secrets:check                       # static leak guard (also runs in CI)
```

`task secrets:encrypt ENV=development FROM=.env` (re)encrypts a flat env file
into `secrets/development.sops.env`. The set is a flat `KEY=VALUE` file: the
SOPS dotenv codec rejects comments and blank lines, so keep the documentation in
`secrets/<environment>.env.example`. CI never decrypts.

## Related

- [rotate-discord-secret.md](rotate-discord-secret.md)
- [rotate-azeroth-soap-credentials.md](rotate-azeroth-soap-credentials.md)
- [../security.md](../security.md)
- [../plan/secrets/README.md](../plan/secrets/README.md)
