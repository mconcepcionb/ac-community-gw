# Retire legacy routes

## Goal

Remove the flat routes and transitional navigation so only the two surfaces
remain, and record the decision and documentation.

## Context

Once 002-008 have relocated every area, the old paths, the home card grid and the
transitional nav are dead code.

## Requirements

- Delete the retired routes and their nav entries: `/accounts`, `/account-links`,
  the flat `/characters` staff view, `/items`, `/identity/users`, the admin
  `/store/products` and `/store/wallet` placement, and any flat `/admin/*` page
  that was rehomed.
- Confirm no route is orphaned and every use case has exactly one home.
- Document the two surfaces, landing, navigation and testing in
  `docs/frontend.md`.
- Write ADR 0013 recording: the two-surface portal/console split,
  permission-driven landing, and the decision to break rather than redirect old
  paths.
- Update `docs/README.md` and `docs/plan/spa/README.md` to point at this plan.

## Acceptance criteria

- No retired route remains; the router tree contains only portal and console
  routes.
- Every replacement route is reachable and permission-gated.
- `task check` and `task web:check` green.

## Implementation notes

- Remove dead components and hooks left behind by the moves; keep shared ones.

## Tests

- Route smoke tests asserting the retired paths no longer resolve.

## Dependencies

- 002-008.
