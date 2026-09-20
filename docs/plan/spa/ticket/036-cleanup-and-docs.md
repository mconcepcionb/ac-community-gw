# Cleanup and docs

## Goal

Remove the legacy test frontend, document the SPA, and enforce the full check
suite.

## Context

Once the SPA covers the auth flow and the shipped features, the throwaway
frontend and the transitional `openapi:check` gaps can be closed.

## Requirements

- Delete `web/legacy/` and any references to it; `ACGW_WEB_DIR` examples point
  to `./web/dist`.
- `task check` runs Go checks plus `openapi:check` (spec + client drift) plus
  `web:check`.
- New `docs/frontend.md` (structure, toolchain, codegen, auth, testing) and an
  ADR recording: OpenAPI generated from Go annotations, generated TS client,
  server cookie sessions for the SPA, and the SPA fallback.
- Update `README.md` (repository layout, quickstart for the SPA) and
  `docs/development.md`.
- Refresh `docs/plan/spa/README.md` status table to `implemented`.

## Acceptance criteria

- `task check` green from a clean checkout (after install/codegen).
- No `web/legacy` references remain.
- Docs describe the full local and production workflow.

## Dependencies

- All previous tickets.
