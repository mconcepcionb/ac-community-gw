# Deploy host and documentation

**Milestone:** D · **Gate:** G-standard · **Depends on:** 002, 003

## Goal

Finalize the deploy-host recipient and the production secret set, and make the
docs describe the shipped SOPS flow.

## Context

S2 leaves the deploy-host identity as a placeholder and there is no
`production.sops.env`. Production secrets are gateway/Discord/SOAP/DB values
specific to the deployment; the deploy host decrypts only its own file.

## Requirements

- Finalize the deploy-host identity and add it to the `production.sops.env`
  creation rule.
- Add `secrets/production.env.example` and
  `secrets/production.sops.env` (encrypted for admin + deploy host).
- Wire the deploy path: document (or add a task/runbook) how the deploy host
  materializes secrets and starts the stack; update
  `docs/runbooks/deploy.md`.
- Update `docs/security.md`, `docs/development.md`, `docs/README.md` and the
  [plan index](../../README.md) to reference the flow.
- Mark this plan delivered; lift any remaining follow-ups into the roadmap.

## Tests

- On a production-like host, `sops decrypt` succeeds only for the files whose
  recipient it holds; the dev host cannot decrypt `production.sops.env`.
- `task secrets:check` green.

## Acceptance criteria

- Deploy host decrypts only its own environment.
- Docs match the shipped behavior; ADR 0015 and the runbooks are consistent.

## Out of scope

- Hosted secret managers; CI decryption.
