# Runbook: deploy

## Purpose

Deploy a new version of `ac-community-gw`.

## Preconditions

- Release image built and pushed, or source at the target revision.
- Target database reachable and backed up.
- The deploy host holds the `deploy-host` age identity and
  `SOPS_AGE_KEY_FILE` points at it, so it can decrypt
  `secrets/production.sops.env` (see
  [secret-management.md](secret-management.md)). The secret set itself is
  created once with `task secrets:encrypt ENV=production FROM=<flat-env-file>`
  from [secrets/production.env.example](../../secrets/production.env.example).
- SOAP endpoint reachable from the private network only.

## Procedure

The gateway and the SPA are separate images served on one origin by Caddy
(`web/Dockerfile`). See [ADR 0012](../ADR/0012-decoupled-spa-serving.md).

```bash
# 1. materialize production secrets on the deploy host
export SOPS_AGE_KEY_FILE=~/.config/ac-community-gw/age/deploy-host.key.txt
task secrets:decrypt ENV=production      # writes .env for Compose

# 2. build and start
task docker:build   # gateway image
task web:image      # SPA + Caddy proxy image
# apply migrations before switching traffic
task db:migrate
# start the new version through your orchestrator or:
task docker:up
```

In production set `ACGW_SITE_ADDRESS` to the public hostname so Caddy obtains
TLS automatically; the gateway must not be exposed directly.

## Validation

```bash
curl https://<host>/healthz
curl https://<host>/readyz
```

Both must return 200. Verify the startup log lists the expected plugins,
commands and services.

## Rollback

See [rollback.md](rollback.md).

## Troubleshooting

- readiness fails: check `readyz` per-check output and database connectivity.
- startup fails: check configuration validation errors in the logs.
