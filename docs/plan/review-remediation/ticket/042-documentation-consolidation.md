# 042 — Documentation consolidation

**Phase:** 7 · **Gate:** G7-docs · **Depends on:** all previous tickets

## Goal

Bring the stable documentation, ADRs, runbooks and README in line with the
remediated behaviour, and record the decisions this plan made.

## Findings addressed

- Documentation drift: `docs/security.md` claims a metrics warning that did not
  exist; the session/role model changed (ticket 010); domain events were resolved
  (ticket 023); command-safety and order-lifecycle rules changed (tickets 002,
  003, 008, 027); metrics/CSP behaviour changed (tickets 013, 039).
- Missing runbooks: secret rotation for the leaked `.env` values, store delivery
  reconciliation, trusted-proxy configuration.

## Context

`docs/*.md`, `docs/ADR/`, `docs/runbooks/`, `README.md`. Per
`docs/README.md`, delivered plan knowledge is consolidated into stable docs and
decisions into ADRs.

## Atomic change

One documentation pass. No code changes.

## Requirements

- Update stable docs to match shipped behaviour:
  - `security.md`: metrics auth requirement, command-safety policy, trusted
    proxies, CSP/headers, `.env` handling.
  - `authentication.md`: role-freshness bound and strategy.
  - `azerothcore-integration.md`: the shared command-safety helpers and the
    executor CR/LF rejection.
  - `store.md`: terminal order state machine, refund-idempotency, reconciliation.
  - `database.md`: new constraints/indexes/partitions.
  - `architecture.md`: resolved event inventory or its removal.
  - `frontend.md`: auth error handling, CSP, metrics routing.
  - `development.md`: new tasks (`test:integration`, coverage, lint).
- Add/extend runbooks:
  - `rotate-discord-secret.md` (rotate the exposed secret).
  - `rotate-azeroth-soap-credentials.md` (verify still accurate).
  - store delivery reconciliation.
  - trusted-proxy configuration.
- Add ADRs for decisions taken:
  - command-safety helper placement and policy;
  - authorization freshness strategy (ticket 010);
  - order refund idempotency/state machine;
  - metrics exposure policy.
- Update `README.md` feature/task tables and `.env.example` for every new key.
- Add this plan's status to `docs/plan/review-remediation/README.md` (mark
  delivered) and note the superseding knowledge pointers.

## Tests

- Link checker over `docs/**` and `README.md` (or a manual pass) reports no broken
  links.
- Every new config key from tickets 006, 011, 012, 013, 014, 018, 029 appears in
  `.env.example`.
- Each ADR follows the existing ADR template/numbering.

## Acceptance criteria (gate G7-docs)

- All doc links resolve.
- Documented behaviour matches the shipped behaviour (spot-checked against code).
- ADRs are numbered consistently and referenced from stable docs.
- The remediation plan is marked delivered.

## Rollback

Revert the docs.

## Out of scope

- Marketing/landing copy.
- Rewriting the SPA's inline help text.
