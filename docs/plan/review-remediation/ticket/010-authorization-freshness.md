# 010 — Authorization freshness and role revocation

**Phase:** 2 · **Gate:** G-standard, contributes to G2-correctness · **Depends on:** 001

## Goal

Ensure that a role change or user ban in Discord takes effect promptly instead of
persisting until the session expires.

## Findings addressed

- Medium: `core/auth/session.go` stores `Roles` at session creation and returns
  them verbatim on `Resolve`; core authorizes on `principal.Roles`. There is no
  re-resolution or invalidation hook, so a demoted/banned user keeps elevated
  permissions for up to `ACGW_SESSION_TTL` (default 24h).

## Context

Roles are synchronized on OAuth callback only. The `identity-discord` plugin is
the owner of sessions and role mappings. A core-level revocation primitive does
not exist.

## Atomic change

Add a session revocation capability keyed by user, and refresh or revoke sessions
when roles change. Choose the cheaper of the two strategies and document it.

## Requirements

- Add `SessionStore.RevokeByUser(ctx, userID)` (memory and Postgres
  implementations) and expose it through `auth.Manager` (or a small interface the
  plugin consumes).
- On role synchronization (`identitydiscord/roles.go`) and on the Discord
  roles-changed event, revoke the affected user's sessions so the next request
  re-authenticates, or refresh the stored roles on each `Resolve`.
- Preferred: resolve roles per request from the authorizer/role store with a
  short cache TTL, avoiding forced re-login. If that is not feasible, revoke on
  change. Record the decision in an ADR.
- Sessions of a banned/demoted user must stop granting the removed permissions no
  later than the chosen freshness bound; document the bound.
- Preserve the existing behavior that a fresh token is minted on every login.

## Tests

- Test: create a session with role `admin`, revoke by user, then `Resolve`
  returns not-found/anonymous.
- Test: with the refresh strategy, changing the role grant changes the principal's
  effective permissions on the next request without re-login.
- Integration test (Postgres) for `RevokeByUser` deleting all sessions of a user.
- Event test: the roles-changed event triggers the chosen invalidation.

## Acceptance criteria (gate G-standard, G2-correctness)

- `task check` green; integration tests pass.
- A revoked/demoted user cannot perform a privileged action after the documented
  freshness bound.
- The chosen strategy and bound are documented; an ADR is added.

## Rollback

Revert. This reopens the revocation lag; note it in the ADR if reverted.

## Out of scope

- Changing role mapping sources (config vs table).
- Ticket 017, which only cleans up session cookie/store details.
