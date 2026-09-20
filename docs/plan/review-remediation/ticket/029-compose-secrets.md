# 029 — Compose exposure and secret handling

**Phase:** 5 · **Gate:** G-standard, contributes to G5-infra · **Depends on:** 012

## Goal

Stop publishing database ports on all interfaces with known credentials, stop
echoing secrets through Compose interpolation, and fix the container-to-container
DSN defaults.

## Findings addressed

- High: `compose.yaml` publishes `5432:5432` and `3306:3306` on all interfaces
  with `acgw/acgw` and `root/root` credentials.
- Medium: `.env` holds a live-looking Discord client secret; `docker compose
  config` echoes interpolated secrets to stdout/logs.
- Nit: the app's AzerothCore DSNs resolve to `localhost` inside the app container
  instead of the `mariadb` service.
- Nit: no restart policy for the stateful/app services.

## Context

`compose.yaml`, `.dockerignore`, `.gitignore`, `.env.example`,
`docs/runbooks/*`. `.env` is correctly gitignored but still present on disk.

## Atomic change

Bind database ports to loopback (or remove them), move credentials to required
variables without defaults, and separate host-oriented `.env` values from
container overrides.

## Requirements

- Bind `5432`/`3306` to `127.0.0.1` only, or drop the `ports:` mapping and use
  `expose:` for inter-container traffic.
- Remove default credentials; require them via environment without a fallback so
  `docker compose up` fails loudly when unset (development values documented in
  `.env.example`).
- Provide compose-only overrides so `ACGW_AZEROTH_*_DSN` points at the `mariadb`
  service inside the network, while host usage keeps `localhost`.
- Add `restart: unless-stopped` for stateful/app services.
- Add a secret-rotation runbook (`docs/runbooks/rotate-discord-secret.md` exists;
  extend/add) and document that `.env` is never an artifact/backup target.
- Rotate the exposed Discord secret and SOAP password outside this ticket (an
  operational step, recorded in the ticket description).
- Ensure `.env` remains gitignored and is excluded from the Docker build context
  (ticket 001/032 verify).

## Tests

- `docker compose config` shows no database port bound to `0.0.0.0` and no
  default credential.
- `docker compose up` with missing credentials fails with a clear message.
- From inside the app container, the AzerothCore DSNs resolve to `mariadb`.
- Confirm `.env` is absent from the image build context.

## Acceptance criteria (gate G-standard, G5-infra)

- `task docker:up` smoke passes with explicitly provided credentials.
- No known-credential database is reachable from a non-loopback interface.
- `docker compose config` does not print real secret values.

## Rollback

Revert the compose changes; this re-exposes the databases. Prefer fixing forward.

## Out of scope

- Production secret manager integration (documented as a follow-up, not built).
- Container image hardening (ticket 032).
