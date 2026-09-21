# Development and Compose flow

**Milestone:** B · **Gate:** G-standard · **Depends on:** 003

## Goal

Make the local and Compose workflows consume the decrypted secret set without
changing the environment-variable contract.

## Context

`task run` loads `.env` via Task `dotenv` (`Taskfile.yml:3`). Compose requires
DB passwords with no default but treats the Discord secret and SOAP password as
optional empty defaults (`compose.yaml:52,59`). SOPS protects the committed
artifact; it does not stop `docker compose config` from echoing interpolated
values.

## Requirements

- `task secrets:decrypt ENV=development` writes `.env` (gitignored) from the
  encrypted set, so the existing `dotenv` flow is unchanged.
- `task secrets:materialize ENV=development` writes the decrypted file to the
  location Compose consumes, with restrictive permissions.
- Update `docs/development.md` setup: replace `cp .env.example .env` with the
  decrypt step (keep the copy path documented as the no-SOPS fallback).
- Update `.env.example` comments to point at the `secrets/` flow.
- Clarify the Compose boundary: document that runtime interpolation is not
  protected by SOPS, and consider making the Discord secret/SOAP password
  required in production configuration (not in dev Compose).

## Tests

- A fresh clone plus the admin age key brings up `task run` and
  `docker compose up`.
- `task secrets:check` stays green; no plaintext secret is committed.

## Acceptance criteria

- Local dev and Compose work from decrypted materialized values.
- `docs/development.md` documents both the SOPS path and the fallback.

## Out of scope

- CI (S5), production deploy (S7).
