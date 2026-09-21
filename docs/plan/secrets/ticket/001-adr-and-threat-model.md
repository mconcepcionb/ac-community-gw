# ADR and threat model

**Milestone:** A · **Gate:** G-standard · **Depends on:** —

## Goal

Record the decision to manage secrets with SOPS + age, define the trust domains
and recipients, and state that CI stays secret-free.

## Context

The repo has no secret manager. The next free ADR number is **0015**; the
[secrets plan](../README.md) reserves `0015-sops-age-secrets.md`. The sibling
`homelab-config` repo runs the same model but with its own identities.

## Requirements

- Write `docs/ADR/0015-sops-age-secrets.md` following the existing ADR structure
  (`## Status`, `## Context`, `## Decision`, `## Consequences`, `## See also`).
- Define the trust domains: dedicated **admin** identity, **dev-host** identity,
  **deploy-host** identity; CI holds none.
- State what is encrypted (`secrets/*.sops.env`) versus what stays out
  (templates only; GitHub Actions handles any CI-only value).
- Document how the admin identity is recovered/stored (outside the repo).
- Reference the reused `homelab-config` model and why identities are **not**
  shared.
- Link the ADR from `docs/security.md` and `docs/development.md`.

## Tests

- None (documentation). Link check: no broken references.

## Acceptance criteria

- ADR 0015 merged and referenced from the secrets plan and security docs.
- Recipients and the CI-secret-free decision are unambiguous.

## Out of scope

- Creating the keys (S2) and any `.sops.yaml` content beyond the decision.
