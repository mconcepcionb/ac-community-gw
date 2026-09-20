# Runbook: deploy

## Purpose

Deploy a new version of `ac-community-gw`.

## Preconditions

- Release image built and pushed, or source at the target revision.
- Target database reachable and backed up.
- Required environment variables available to the runtime.
- SOAP endpoint reachable from the private network only.

## Procedure

The gateway and the SPA are separate images served on one origin by Caddy
(`web/Dockerfile`). See [ADR 0012](../ADR/0012-decoupled-spa-serving.md).

```bash
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
